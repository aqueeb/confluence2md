package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsConfluenceMIME(t *testing.T) {
	// Create a temp file with valid Confluence MIME headers
	validContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body>Test</body></html>
------=_Part_123_456789.123456789--
`
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "valid.doc")
	if err := os.WriteFile(validFile, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test valid Confluence MIME
	isConfluence, err := IsConfluenceMIME(validFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !isConfluence {
		t.Error("Expected valid Confluence MIME file to return true")
	}

	// Create an invalid file (plain text)
	invalidContent := `This is just plain text.
Not a MIME message at all.
`
	invalidFile := filepath.Join(tmpDir, "invalid.doc")
	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test invalid file
	isConfluence, err = IsConfluenceMIME(invalidFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if isConfluence {
		t.Error("Expected invalid file to return false")
	}

	// Test non-existent file (should return error)
	isConfluence, err = IsConfluenceMIME("/nonexistent/file.doc")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if isConfluence {
		t.Error("Expected non-existent file to return false")
	}
}

func TestExtractHTMLFromMIME(t *testing.T) {
	// Create a temp file with valid Confluence MIME content
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><head><title>Test</title></head><body><h1>Hello World</h1></body></html>
------=_Part_123_456789.123456789--
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.doc")
	if err := os.WriteFile(testFile, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Extract HTML
	html, err := ExtractHTMLFromMIME(testFile)
	if err != nil {
		t.Fatalf("ExtractHTMLFromMIME failed: %v", err)
	}

	// Verify HTML content
	if html == "" {
		t.Error("Expected non-empty HTML content")
	}
	if !strings.Contains(html, "<h1>Hello World</h1>") {
		t.Errorf("Expected HTML to contain '<h1>Hello World</h1>', got: %s", html)
	}
}

func TestExtractHTMLFromMIME_QuotedPrintable(t *testing.T) {
	// Test with quoted-printable encoded content
	// "=" at end of line means soft line break, "=3D" means "="
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body>Test =3D Value</body></html>
------=_Part_123_456789.123456789--
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.doc")
	if err := os.WriteFile(testFile, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	html, err := ExtractHTMLFromMIME(testFile)
	if err != nil {
		t.Fatalf("ExtractHTMLFromMIME failed: %v", err)
	}

	// Verify quoted-printable was decoded (=3D should become =)
	if !strings.Contains(html, "Test = Value") {
		t.Errorf("Expected decoded '=' in HTML, got: %s", html)
	}
}

func TestExtractHTMLFromMIME_Errors(t *testing.T) {
	// Test non-existent file
	_, err := ExtractHTMLFromMIME("/nonexistent/file.doc")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test file without HTML part
	noHTMLContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/plain; charset=UTF-8

Just plain text, no HTML
------=_Part_123_456789.123456789--
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-html.doc")
	if err := os.WriteFile(testFile, []byte(noHTMLContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = ExtractHTMLFromMIME(testFile)
	if err == nil {
		t.Error("Expected error for file without HTML part")
	}
}

func TestExtractHTMLFromMIME_InvalidMIME(t *testing.T) {
	tmpDir := t.TempDir()

	// Test non-multipart content type
	nonMultipartContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Test
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

This is not a multipart message.
`
	nonMultipartFile := filepath.Join(tmpDir, "non-multipart.doc")
	if err := os.WriteFile(nonMultipartFile, []byte(nonMultipartContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ExtractHTMLFromMIME(nonMultipartFile)
	if err == nil {
		t.Error("Expected error for non-multipart content")
	}
	if !strings.Contains(err.Error(), "expected multipart") {
		t.Errorf("Expected 'expected multipart' error, got: %v", err)
	}

	// Test multipart without boundary
	noBoundaryContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Test
MIME-Version: 1.0
Content-Type: multipart/related

This has no boundary.
`
	noBoundaryFile := filepath.Join(tmpDir, "no-boundary.doc")
	if err := os.WriteFile(noBoundaryFile, []byte(noBoundaryContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = ExtractHTMLFromMIME(noBoundaryFile)
	if err == nil {
		t.Error("Expected error for missing boundary")
	}
	if !strings.Contains(err.Error(), "no boundary") {
		t.Errorf("Expected 'no boundary' error, got: %v", err)
	}

	// Test invalid MIME message (not parseable)
	invalidMIMEContent := `This is not a valid MIME message at all.
No headers, no structure.
`
	invalidMIMEFile := filepath.Join(tmpDir, "invalid-mime.doc")
	if err := os.WriteFile(invalidMIMEFile, []byte(invalidMIMEContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = ExtractHTMLFromMIME(invalidMIMEFile)
	if err == nil {
		t.Error("Expected error for invalid MIME message")
	}
}

func TestExtractHTMLFromMIME_NoTransferEncoding(t *testing.T) {
	// Test HTML without transfer encoding (should read directly)
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8

<html><body><h1>Direct Content</h1></body></html>
------=_Part_123_456789.123456789--
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-encoding.doc")
	if err := os.WriteFile(testFile, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	html, err := ExtractHTMLFromMIME(testFile)
	if err != nil {
		t.Fatalf("ExtractHTMLFromMIME failed: %v", err)
	}

	if !strings.Contains(html, "Direct Content") {
		t.Errorf("Expected HTML content, got: %s", html)
	}
}

func TestExtractHTMLFromMIME_MultipleParts(t *testing.T) {
	// Test MIME with multiple parts (image and HTML)
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: image/png
Content-Transfer-Encoding: base64

iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==
------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>After Image Part</h1></body></html>
------=_Part_123_456789.123456789--
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "multi-part.doc")
	if err := os.WriteFile(testFile, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	html, err := ExtractHTMLFromMIME(testFile)
	if err != nil {
		t.Fatalf("ExtractHTMLFromMIME failed: %v", err)
	}

	if !strings.Contains(html, "After Image Part") {
		t.Errorf("Expected HTML content from second part, got: %s", html)
	}
}

func TestIsConfluenceMIME_PartialHeaders(t *testing.T) {
	tmpDir := t.TempDir()

	// Test file with only Date header (missing MIME-Version and Subject)
	onlyDateContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Content-Type: text/plain

Some content
`
	onlyDateFile := filepath.Join(tmpDir, "only-date.doc")
	if err := os.WriteFile(onlyDateFile, []byte(onlyDateContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	isConfluence, err := IsConfluenceMIME(onlyDateFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if isConfluence {
		t.Error("Expected file with only Date header to return false")
	}

	// Test file with Date and MIME-Version but missing Confluence subject
	missingSubjectContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
MIME-Version: 1.0
Content-Type: text/plain

Some content
`
	missingSubjectFile := filepath.Join(tmpDir, "missing-subject.doc")
	if err := os.WriteFile(missingSubjectFile, []byte(missingSubjectContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	isConfluence, err = IsConfluenceMIME(missingSubjectFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if isConfluence {
		t.Error("Expected file without Confluence subject to return false")
	}

	// Test file with Subject but wrong content
	wrongSubjectContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
MIME-Version: 1.0
Subject: Random Email Subject

Some content
`
	wrongSubjectFile := filepath.Join(tmpDir, "wrong-subject.doc")
	if err := os.WriteFile(wrongSubjectFile, []byte(wrongSubjectContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	isConfluence, err = IsConfluenceMIME(wrongSubjectFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if isConfluence {
		t.Error("Expected file with wrong subject to return false")
	}
}

