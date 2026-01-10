package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aqueeb/confluence2md/converter"
)

func TestGenerateOutputPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic doc to md",
			input:    "file.doc",
			expected: "file.md",
		},
		{
			name:     "plus to dash in filename",
			input:    "my+file.doc",
			expected: "my-file.md",
		},
		{
			name:     "multiple plus signs",
			input:    "my+complex+file+name.doc",
			expected: "my-complex-file-name.md",
		},
		{
			name:     "with directory path",
			input:    "/path/to/docs/file.doc",
			expected: "/path/to/docs/file.md",
		},
		{
			name:     "with directory and plus",
			input:    "/path/to/docs/my+file.doc",
			expected: "/path/to/docs/my-file.md",
		},
		{
			name:     "relative path",
			input:    "./docs/file.doc",
			expected: "docs/file.md",
		},
		{
			name:     "no extension (edge case)",
			input:    "filename",
			expected: "filename.md",
		},
		{
			name:     "already md extension kept as is",
			input:    "file.md.doc",
			expected: "file.md.md",
		},
		{
			name:     "multiple dots in filename",
			input:    "file.backup.doc",
			expected: "file.backup.md",
		},
		{
			name:     "uppercase DOC extension",
			input:    "file.DOC",
			expected: "file.DOC.md", // TrimSuffix is case-sensitive
		},
		{
			name:     "plus in directory path preserved",
			input:    "/path+dir/file.doc",
			expected: "/path+dir/file.md", // only filename is modified
		},
		{
			name:     "empty filename with doc extension",
			input:    ".doc",
			expected: ".md",
		},
		{
			name:     "spaces in filename",
			input:    "my file name.doc",
			expected: "my file name.md",
		},
		{
			name:     "unicode in filename",
			input:    "文档.doc",
			expected: "文档.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateOutputPath(tt.input)
			if got != tt.expected {
				t.Errorf("generateOutputPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// createTestConfluenceMIME creates a valid Confluence MIME file for testing
func createTestConfluenceMIME(t *testing.T, dir, filename, htmlContent string) string {
	t.Helper()
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

` + htmlContent + `
------=_Part_123_456789.123456789--
`
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// createPlainTextFile creates a non-MIME file for testing
func createPlainTextFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

func TestConvertFile_DryRun(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")
	outputPath := filepath.Join(tmpDir, "test.md")

	// Run in dry-run mode
	err := convertFile(inputPath, outputPath, false, true)
	if err != nil {
		t.Fatalf("convertFile dry-run failed: %v", err)
	}

	// Verify output file was NOT created
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Expected output file to NOT exist in dry-run mode")
	}
}

func TestConvertFile_Success(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Hello World</h1><p>This is a test.</p></body></html>")
	outputPath := filepath.Join(tmpDir, "test.md")

	// Run conversion
	err := convertFile(inputPath, outputPath, false, false)
	if err != nil {
		t.Fatalf("convertFile failed: %v", err)
	}

	// Verify output file exists
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify content
	md := string(content)
	if !strings.Contains(md, "Hello World") {
		t.Errorf("Expected markdown to contain 'Hello World', got: %s", md)
	}
	if !strings.Contains(md, "This is a test") {
		t.Errorf("Expected markdown to contain 'This is a test', got: %s", md)
	}
}

func TestConvertFile_NonExistentInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "nonexistent.doc")
	outputPath := filepath.Join(tmpDir, "output.md")

	err := convertFile(inputPath, outputPath, false, false)
	if err == nil {
		t.Error("Expected error for non-existent input file")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestConvertFile_InvalidMIME(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createPlainTextFile(t, tmpDir, "invalid.doc", "This is just plain text, not MIME.")
	outputPath := filepath.Join(tmpDir, "invalid.md")

	err := convertFile(inputPath, outputPath, false, false)
	if err == nil {
		t.Error("Expected error for non-MIME file")
	}
	if !strings.Contains(err.Error(), "does not appear to be a Confluence MIME export") {
		t.Errorf("Expected MIME validation error, got: %v", err)
	}
}

func TestConvertFile_VerboseMode(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")
	outputPath := filepath.Join(tmpDir, "test.md")

	// Verbose mode should not cause errors
	err := convertFile(inputPath, outputPath, true, false)
	if err != nil {
		t.Fatalf("convertFile with verbose failed: %v", err)
	}

	// Verify output was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected output file to exist")
	}
}

func TestConvertDirectory_MultipleFiles(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()

	// Create multiple Confluence MIME files
	createTestConfluenceMIME(t, tmpDir, "doc1.doc", "<html><body><h1>Doc 1</h1></body></html>")
	createTestConfluenceMIME(t, tmpDir, "doc2.doc", "<html><body><h1>Doc 2</h1></body></html>")
	createTestConfluenceMIME(t, tmpDir, "doc3.doc", "<html><body><h1>Doc 3</h1></body></html>")

	// Run directory conversion
	err := convertDirectory(tmpDir, false, false)
	if err != nil {
		t.Fatalf("convertDirectory failed: %v", err)
	}

	// Verify all output files exist
	for _, name := range []string{"doc1.md", "doc2.md", "doc3.md"} {
		outputPath := filepath.Join(tmpDir, name)
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s to exist", name)
		}
	}
}

func TestConvertDirectory_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Run directory conversion on empty directory
	err := convertDirectory(tmpDir, false, false)
	if err != nil {
		t.Fatalf("convertDirectory on empty dir failed: %v", err)
	}
	// Should complete without error, just print "No .doc files found"
}

func TestConvertDirectory_MixedFiles(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()

	// Create Confluence MIME file
	createTestConfluenceMIME(t, tmpDir, "valid.doc", "<html><body><h1>Valid</h1></body></html>")

	// Create non-MIME .doc file
	createPlainTextFile(t, tmpDir, "invalid.doc", "This is not a MIME file")

	// Create other file types (should be ignored)
	createPlainTextFile(t, tmpDir, "readme.txt", "Just a text file")
	createPlainTextFile(t, tmpDir, "data.json", "{}")

	// Run directory conversion
	err := convertDirectory(tmpDir, false, false)
	if err != nil {
		t.Fatalf("convertDirectory failed: %v", err)
	}

	// Only valid.md should exist
	validOutput := filepath.Join(tmpDir, "valid.md")
	if _, err := os.Stat(validOutput); os.IsNotExist(err) {
		t.Error("Expected valid.md to exist")
	}

	// invalid.md should NOT exist (file was not valid MIME)
	invalidOutput := filepath.Join(tmpDir, "invalid.md")
	if _, err := os.Stat(invalidOutput); !os.IsNotExist(err) {
		t.Error("Expected invalid.md to NOT exist (source was not Confluence MIME)")
	}

	// Other file types should not have .md counterparts
	txtOutput := filepath.Join(tmpDir, "readme.md")
	if _, err := os.Stat(txtOutput); !os.IsNotExist(err) {
		t.Error("Expected readme.md to NOT exist (source was .txt)")
	}
}

func TestConvertDirectory_DryRun(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()

	// Create Confluence MIME files
	createTestConfluenceMIME(t, tmpDir, "doc1.doc", "<html><body><h1>Doc 1</h1></body></html>")
	createTestConfluenceMIME(t, tmpDir, "doc2.doc", "<html><body><h1>Doc 2</h1></body></html>")

	// Run directory conversion in dry-run mode
	err := convertDirectory(tmpDir, false, true)
	if err != nil {
		t.Fatalf("convertDirectory dry-run failed: %v", err)
	}

	// Verify NO output files were created
	for _, name := range []string{"doc1.md", "doc2.md"} {
		outputPath := filepath.Join(tmpDir, name)
		if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
			t.Errorf("Expected output file %s to NOT exist in dry-run mode", name)
		}
	}
}

func TestConvertDirectory_NoConfluenceFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only non-MIME .doc files
	createPlainTextFile(t, tmpDir, "plain1.doc", "Plain text 1")
	createPlainTextFile(t, tmpDir, "plain2.doc", "Plain text 2")

	// Run directory conversion
	err := convertDirectory(tmpDir, false, false)
	if err != nil {
		t.Fatalf("convertDirectory failed: %v", err)
	}
	// Should complete without error, prints "No Confluence MIME exports found"
}

func TestConvertDirectory_VerboseMode(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()

	// Create mix of files
	createTestConfluenceMIME(t, tmpDir, "valid.doc", "<html><body><h1>Valid</h1></body></html>")
	createPlainTextFile(t, tmpDir, "invalid.doc", "Not MIME")

	// Verbose mode should not cause errors
	err := convertDirectory(tmpDir, true, false)
	if err != nil {
		t.Fatalf("convertDirectory with verbose failed: %v", err)
	}

	// Verify valid.md exists
	if _, err := os.Stat(filepath.Join(tmpDir, "valid.md")); os.IsNotExist(err) {
		t.Error("Expected valid.md to exist")
	}
}

func TestConvertDirectory_NonExistentDirectory(t *testing.T) {
	err := convertDirectory("/nonexistent/directory/path", false, false)
	if err != nil {
		// filepath.Glob doesn't error on non-existent paths, it just returns empty
		// So this should not error, but print "No .doc files found"
		t.Logf("Got error (may be expected depending on implementation): %v", err)
	}
}

func TestConvertDirectory_WithPlusInFilename(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()

	// Create file with + in name
	createTestConfluenceMIME(t, tmpDir, "my+doc+file.doc", "<html><body><h1>Plus Test</h1></body></html>")

	// Run directory conversion
	err := convertDirectory(tmpDir, false, false)
	if err != nil {
		t.Fatalf("convertDirectory failed: %v", err)
	}

	// Verify output file has - instead of +
	outputPath := filepath.Join(tmpDir, "my-doc-file.md")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected my-doc-file.md to exist (+ should be converted to -)")
	}

	// Original + filename should not have .md version
	wrongPath := filepath.Join(tmpDir, "my+doc+file.md")
	if _, err := os.Stat(wrongPath); !os.IsNotExist(err) {
		t.Error("Did not expect my+doc+file.md to exist")
	}
}
