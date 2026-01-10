package pandoc

import (
	"context"
	"os"
	"runtime"
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

func TestGetBinaryName(t *testing.T) {
	name := getBinaryName()
	if runtime.GOOS == "windows" {
		if name != "pandoc.exe" {
			t.Errorf("expected 'pandoc.exe' on Windows, got: %s", name)
		}
	} else {
		if name != "pandoc" {
			t.Errorf("expected 'pandoc' on non-Windows, got: %s", name)
		}
	}
}

func TestCleanup(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	// Ensure extracted first
	path, err := EnsureExtracted()
	if err != nil {
		t.Fatalf("EnsureExtracted failed: %v", err)
	}

	// Verify the binary exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected extracted binary to exist")
	}

	// Run cleanup
	err = Cleanup()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify GetPath returns empty after cleanup
	if GetPath() != "" {
		t.Error("GetPath should return empty string after Cleanup")
	}

	// Re-extract for subsequent tests
	_, err = EnsureExtracted()
	if err != nil {
		t.Fatalf("Re-extraction after cleanup failed: %v", err)
	}
}

func TestCleanupWhenNotExtracted(t *testing.T) {
	// Save current state
	oldPath := extractedPath

	// Simulate not extracted state
	extractedPath = ""

	// Cleanup should succeed and do nothing
	err := Cleanup()
	if err != nil {
		t.Errorf("Cleanup when not extracted should not error: %v", err)
	}

	// Restore state
	extractedPath = oldPath
}

func TestConvertWithExtraArgs(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test conversion with multiple extra args
	input := []byte("<table><tr><td>Cell 1</td><td>Cell 2</td></tr></table>")
	output, err := Convert(ctx, input, "html", "gfm", "--wrap=none", "--columns=1000")
	if err != nil {
		t.Fatalf("Convert with extra args failed: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "Cell 1") || !strings.Contains(result, "Cell 2") {
		t.Errorf("output doesn't contain expected cells: %s", result)
	}
}

func TestRunWithInvalidArgs(t *testing.T) {
	if !IsEmbedded() {
		t.Skip("pandoc binary not embedded (run scripts/download-pandoc.sh first)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run with help flag (should succeed)
	output, err := Run(ctx, "--help")
	if err != nil {
		t.Fatalf("Run --help failed: %v", err)
	}

	if !strings.Contains(string(output), "pandoc") {
		t.Errorf("help output doesn't mention pandoc: %s", string(output))
	}
}

func TestEmbeddedSize(t *testing.T) {
	size := EmbeddedSize()
	if IsEmbedded() {
		if size == 0 {
			t.Error("embedded size should not be 0 when embedded")
		}
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
	// Version should be a valid semver-like string
	parts := strings.Split(Version, ".")
	if len(parts) < 2 {
		t.Errorf("Version should have at least 2 parts: %s", Version)
	}
}
