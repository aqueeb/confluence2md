package pandoc

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIsEmbedded(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	size := EmbeddedSize()
	t.Logf("Embedded binary size: %d bytes (%.1f MB)", size, float64(size)/1024/1024)

	// Pandoc binary should be at least 50MB
	if size < 50*1024*1024 {
		t.Errorf("embedded binary seems too small: %d bytes", size)
	}
}

func TestEnsureExtracted(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	path, err := EnsureExtracted()
	if err != nil {
		t.Fatalf("EnsureExtracted failed: %v", err)
	}

	if path == "" {
		t.Fatal("extracted path is empty")
	}

	t.Logf("Extracted to: %s", path)
}

func TestGetVersion(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	version, err := GetVersion(ctx)
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if !strings.Contains(version, "pandoc") {
		t.Errorf("version string doesn't contain 'pandoc': %s", version)
	}

	t.Logf("Pandoc version: %s", version)
}

func TestConvertMarkdownToHTML(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := []byte("# Hello\n\nThis is a **test**.")
	output, err := Convert(ctx, input, "markdown", "html")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "<h1") {
		t.Errorf("output doesn't contain h1 tag: %s", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("output doesn't contain 'Hello': %s", result)
	}
	if !strings.Contains(result, "<strong>") || !strings.Contains(result, "test") {
		t.Errorf("output doesn't contain bold text: %s", result)
	}

	t.Logf("Converted output:\n%s", result)
}

func TestConvertHTMLToGFM(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := []byte("<h1>Title</h1><p>Paragraph with <strong>bold</strong> text.</p>")
	output, err := Convert(ctx, input, "html", "gfm", "--wrap=none")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "# Title") {
		t.Errorf("output doesn't contain markdown heading: %s", result)
	}
	if !strings.Contains(result, "**bold**") {
		t.Errorf("output doesn't contain markdown bold: %s", result)
	}

	t.Logf("Converted output:\n%s", result)
}

func TestConcurrentAccess(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Launch 10 concurrent extractions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := EnsureExtracted()
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent extraction failed: %v", err)
	}
}

func TestRun(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := Run(ctx, "--version")
	if err != nil {
		t.Fatalf("Run --version failed: %v", err)
	}

	if !strings.Contains(string(output), "pandoc") {
		t.Errorf("output doesn't contain 'pandoc': %s", string(output))
	}
}

func TestGetPath(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	// Ensure extracted first
	_, err := EnsureExtracted()
	if err != nil {
		t.Fatalf("EnsureExtracted failed: %v", err)
	}

	path := GetPath()
	if path == "" {
		t.Error("GetPath returned empty string after extraction")
	}
}
