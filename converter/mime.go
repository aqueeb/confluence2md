package converter

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"strings"
)

// ExtractHTMLFromMIME reads a MIME-encoded Confluence export file and extracts the HTML content.
func ExtractHTMLFromMIME(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse as email/MIME message
	msg, err := mail.ReadMessage(bufio.NewReader(file))
	if err != nil {
		return "", fmt.Errorf("failed to parse MIME message: %w", err)
	}

	contentType := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("failed to parse Content-Type: %w", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return "", fmt.Errorf("expected multipart message, got: %s", mediaType)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return "", fmt.Errorf("no boundary found in Content-Type")
	}

	// Parse multipart body
	mr := multipart.NewReader(msg.Body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read MIME part: %w", err)
		}

		partContentType := part.Header.Get("Content-Type")
		partMediaType, _, _ := mime.ParseMediaType(partContentType)

		// We're looking for the text/html part
		if partMediaType == "text/html" {
			encoding := part.Header.Get("Content-Transfer-Encoding")

			var reader io.Reader = part
			if strings.ToLower(encoding) == "quoted-printable" {
				reader = quotedprintable.NewReader(part)
			}

			htmlBytes, err := io.ReadAll(reader)
			if err != nil {
				return "", fmt.Errorf("failed to read HTML content: %w", err)
			}

			return string(htmlBytes), nil
		}
	}

	return "", fmt.Errorf("no text/html part found in MIME message")
}

// IsConfluenceMIME checks if a file appears to be a MIME-encoded Confluence export.
func IsConfluenceMIME(filepath string) bool {
	file, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	hasDateHeader := false
	hasMIMEVersion := false
	hasConfluenceSubject := false

	for scanner.Scan() && lineCount < 10 {
		line := scanner.Text()
		lineCount++

		if strings.HasPrefix(line, "Date:") {
			hasDateHeader = true
		}
		if strings.HasPrefix(line, "MIME-Version:") {
			hasMIMEVersion = true
		}
		if strings.Contains(line, "Exported From Confluence") {
			hasConfluenceSubject = true
		}
	}

	return hasDateHeader && hasMIMEVersion && hasConfluenceSubject
}
