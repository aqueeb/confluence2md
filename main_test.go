package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aqueeb/confluence2md/converter"
)

// Tests for parseFlags function
func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantVersion bool
		wantVerbose bool
		wantDryRun  bool
		wantDir     string
		wantOutput  string
		wantArgs    []string
	}{
		{
			name:        "version flag",
			args:        []string{"--version"},
			wantVersion: true,
		},
		{
			name:       "verbose short flag",
			args:       []string{"-v", "input.doc"},
			wantVerbose: true,
			wantArgs:   []string{"input.doc"},
		},
		{
			name:       "verbose long flag",
			args:       []string{"--verbose", "input.doc"},
			wantVerbose: true,
			wantArgs:   []string{"input.doc"},
		},
		{
			name:       "dry-run flag",
			args:       []string{"--dry-run", "input.doc"},
			wantDryRun: true,
			wantArgs:   []string{"input.doc"},
		},
		{
			name:    "dir flag",
			args:    []string{"--dir", "/path/to/docs"},
			wantDir: "/path/to/docs",
		},
		{
			name:       "output short flag",
			args:       []string{"-o", "output.md", "input.doc"},
			wantOutput: "output.md",
			wantArgs:   []string{"input.doc"},
		},
		{
			name:       "output long flag",
			args:       []string{"--output", "output.md", "input.doc"},
			wantOutput: "output.md",
			wantArgs:   []string{"input.doc"},
		},
		{
			name:        "combined flags",
			args:        []string{"-v", "--dry-run", "-o", "out.md", "in.doc"},
			wantVerbose: true,
			wantDryRun:  true,
			wantOutput:  "out.md",
			wantArgs:    []string{"in.doc"},
		},
		{
			name:     "no flags",
			args:     []string{"input.doc"},
			wantArgs: []string{"input.doc"},
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg, err := parseFlags(tt.args, &buf)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if cfg.showVersion != tt.wantVersion {
				t.Errorf("showVersion = %v, want %v", cfg.showVersion, tt.wantVersion)
			}
			if cfg.verbose != tt.wantVerbose {
				t.Errorf("verbose = %v, want %v", cfg.verbose, tt.wantVerbose)
			}
			if cfg.dryRun != tt.wantDryRun {
				t.Errorf("dryRun = %v, want %v", cfg.dryRun, tt.wantDryRun)
			}
			if cfg.dirMode != tt.wantDir {
				t.Errorf("dirMode = %v, want %v", cfg.dirMode, tt.wantDir)
			}
			if cfg.outputPath != tt.wantOutput {
				t.Errorf("outputPath = %v, want %v", cfg.outputPath, tt.wantOutput)
			}
			if len(cfg.args) != len(tt.wantArgs) {
				t.Errorf("args = %v, want %v", cfg.args, tt.wantArgs)
			}
		})
	}
}

// Tests for run function
func TestRun_Version(t *testing.T) {
	cfg := &config{showVersion: true}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := run(cfg)

	w.Close()
	os.Stdout = old

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "confluence2md") {
		t.Errorf("Expected version output, got: %s", output)
	}
}

func TestRun_NoArgs(t *testing.T) {
	cfg := &config{args: []string{}}

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := run(cfg)

	w.Close()
	os.Stderr = old

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for no args, got %d", exitCode)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage output, got: %s", output)
	}
}

func TestRun_DirMode(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	tmpDir := t.TempDir()
	createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")

	cfg := &config{
		dirMode: tmpDir,
		dryRun:  true, // Use dry-run to avoid star prompt
	}

	exitCode := run(cfg)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_SingleFile(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")

	cfg := &config{
		args:   []string{inputPath},
		dryRun: true,
	}

	exitCode := run(cfg)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_NonExistentFile(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	cfg := &config{
		args: []string{"/nonexistent/file.doc"},
	}

	exitCode := run(cfg)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for non-existent file, got %d", exitCode)
	}
}

func TestRun_CustomOutput(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")
	customOutput := filepath.Join(tmpDir, "custom.md")

	cfg := &config{
		args:       []string{inputPath},
		outputPath: customOutput,
	}

	exitCode := run(cfg)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if _, err := os.Stat(customOutput); os.IsNotExist(err) {
		t.Error("Expected custom output file to be created")
	}
}

// TestMain_Version tests the --version flag using subprocess
func TestMain_Version(t *testing.T) {
	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Run with --version flag
	cmd = exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v\n%s", err, output)
	}

	result := string(output)
	if !strings.Contains(result, "confluence2md") {
		t.Errorf("Expected version output to contain 'confluence2md', got: %s", result)
	}
}

// TestMain_NoArgs tests running without arguments (should show usage)
func TestMain_NoArgs(t *testing.T) {
	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Run without arguments
	cmd = exec.Command(binaryPath)
	output, _ := cmd.CombinedOutput() // Expect non-zero exit code

	result := string(output)
	if !strings.Contains(result, "Usage:") {
		t.Errorf("Expected usage output, got: %s", result)
	}
}

// TestMain_SingleFile tests converting a single file via the binary
func TestMain_SingleFile(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create a test file
	inputPath := filepath.Join(tmpDir, "test.doc")
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123"

------=_Part_123
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>CLI Test</h1></body></html>
------=_Part_123--
`
	if err := os.WriteFile(inputPath, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run conversion
	cmd = exec.Command(binaryPath, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Conversion failed: %v\n%s", err, output)
	}

	// Verify output file exists
	outputPath := filepath.Join(tmpDir, "test.md")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	// Verify output contains expected content
	mdContent, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(mdContent), "CLI Test") {
		t.Errorf("Expected output to contain 'CLI Test', got: %s", mdContent)
	}
}

// TestMain_DirMode tests the --dir flag via the binary
func TestMain_DirMode(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create a subdirectory with test files
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123"

------=_Part_123
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>Dir Mode Test</h1></body></html>
------=_Part_123--
`
	if err := os.WriteFile(filepath.Join(docsDir, "doc1.doc"), []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run with --dir flag
	cmd = exec.Command(binaryPath, "--dir", docsDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--dir conversion failed: %v\n%s", err, output)
	}

	// Verify output
	result := string(output)
	if !strings.Contains(result, "Converted") {
		t.Errorf("Expected 'Converted' message, got: %s", result)
	}

	// Verify output file exists
	if _, err := os.Stat(filepath.Join(docsDir, "doc1.md")); os.IsNotExist(err) {
		t.Error("Expected doc1.md to be created")
	}
}

// TestMain_DryRun tests the --dry-run flag via the binary
func TestMain_DryRun(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create a test file
	inputPath := filepath.Join(tmpDir, "test.doc")
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123"

------=_Part_123
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>Dry Run Test</h1></body></html>
------=_Part_123--
`
	if err := os.WriteFile(inputPath, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run with --dry-run flag
	cmd = exec.Command(binaryPath, "--dry-run", inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--dry-run failed: %v\n%s", err, output)
	}

	result := string(output)
	if !strings.Contains(result, "[dry-run]") {
		t.Errorf("Expected '[dry-run]' message, got: %s", result)
	}

	// Verify output file was NOT created
	outputPath := filepath.Join(tmpDir, "test.md")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Expected output file to NOT be created in dry-run mode")
	}
}

// TestMain_Verbose tests the -v flag via the binary
func TestMain_Verbose(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create a test file
	inputPath := filepath.Join(tmpDir, "test.doc")
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123"

------=_Part_123
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>Verbose Test</h1></body></html>
------=_Part_123--
`
	if err := os.WriteFile(inputPath, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run with -v flag
	cmd = exec.Command(binaryPath, "-v", inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("-v failed: %v\n%s", err, output)
	}

	result := string(output)
	if !strings.Contains(result, "Extracting HTML") {
		t.Errorf("Expected verbose output, got: %s", result)
	}
}

// TestMain_CustomOutput tests the -o flag via the binary
func TestMain_CustomOutput(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available: %v", err)
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "confluence2md")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create a test file
	inputPath := filepath.Join(tmpDir, "test.doc")
	customOutput := filepath.Join(tmpDir, "custom-output.md")
	mimeContent := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123"

------=_Part_123
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>Custom Output Test</h1></body></html>
------=_Part_123--
`
	if err := os.WriteFile(inputPath, []byte(mimeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run with -o flag (flags must come before positional args)
	cmd = exec.Command(binaryPath, "-o", customOutput, inputPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("-o failed: %v", err)
	}

	// Verify custom output file exists
	if _, err := os.Stat(customOutput); os.IsNotExist(err) {
		t.Error("Expected custom output file to be created")
	}

	// Verify default output file was NOT created
	defaultOutput := filepath.Join(tmpDir, "test.md")
	if _, err := os.Stat(defaultOutput); !os.IsNotExist(err) {
		t.Error("Expected default output file to NOT be created when using -o")
	}
}

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

func TestPrintStarPrompt(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printStarPrompt()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify output contains expected text
	if !strings.Contains(output, "Star the repo") {
		t.Errorf("Expected output to contain 'Star the repo', got: %s", output)
	}
	if !strings.Contains(output, "github.com/aqueeb/confluence2md") {
		t.Errorf("Expected output to contain repo URL, got: %s", output)
	}
	if !strings.Contains(output, "Buy me a coffee") {
		t.Errorf("Expected output to contain 'Buy me a coffee', got: %s", output)
	}
}

func TestConvertFile_VerboseMessages(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")
	outputPath := filepath.Join(tmpDir, "test.md")

	// Capture stdout to verify verbose messages
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertFile(inputPath, outputPath, true, false)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertFile with verbose failed: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify verbose messages
	if !strings.Contains(output, "Converting:") {
		t.Errorf("Expected 'Converting:' in verbose output, got: %s", output)
	}
	if !strings.Contains(output, "Extracting HTML") {
		t.Errorf("Expected 'Extracting HTML' in verbose output, got: %s", output)
	}
	if !strings.Contains(output, "Converting HTML to Markdown") {
		t.Errorf("Expected 'Converting HTML to Markdown' in verbose output, got: %s", output)
	}
	if !strings.Contains(output, "Writing output") {
		t.Errorf("Expected 'Writing output' in verbose output, got: %s", output)
	}
}

func TestConvertFile_OutputMessages(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "mydoc.doc", "<html><body><h1>Test</h1></body></html>")
	outputPath := filepath.Join(tmpDir, "mydoc.md")

	// Capture stdout to verify output messages
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertFile(inputPath, outputPath, false, false)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertFile failed: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify non-verbose message
	if !strings.Contains(output, "Converted:") {
		t.Errorf("Expected 'Converted:' message, got: %s", output)
	}
	if !strings.Contains(output, "mydoc.doc") {
		t.Errorf("Expected input filename in output, got: %s", output)
	}
	if !strings.Contains(output, "mydoc.md") {
		t.Errorf("Expected output filename in output, got: %s", output)
	}
}

func TestConvertFile_DryRunMessages(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	inputPath := createTestConfluenceMIME(t, tmpDir, "test.doc", "<html><body><h1>Test</h1></body></html>")
	outputPath := filepath.Join(tmpDir, "test.md")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertFile(inputPath, outputPath, false, true)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertFile dry-run failed: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify dry-run message
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("Expected '[dry-run]' message, got: %s", output)
	}
	if !strings.Contains(output, "Would convert") {
		t.Errorf("Expected 'Would convert' message, got: %s", output)
	}
}

func TestConvertDirectory_Messages(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	createTestConfluenceMIME(t, tmpDir, "doc1.doc", "<html><body><h1>Doc 1</h1></body></html>")
	createTestConfluenceMIME(t, tmpDir, "doc2.doc", "<html><body><h1>Doc 2</h1></body></html>")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertDirectory(tmpDir, false, false)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertDirectory failed: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify output messages
	if !strings.Contains(output, "Found 2 Confluence export(s)") {
		t.Errorf("Expected 'Found 2 Confluence export(s)' message, got: %s", output)
	}
	if !strings.Contains(output, "Converted 2/2 files") {
		t.Errorf("Expected 'Converted 2/2 files' message, got: %s", output)
	}
}

func TestConvertDirectory_VerboseSkipMessages(t *testing.T) {
	if err := converter.CheckPandoc(); err != nil {
		t.Skipf("Pandoc not available, skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	createTestConfluenceMIME(t, tmpDir, "valid.doc", "<html><body><h1>Valid</h1></body></html>")
	createPlainTextFile(t, tmpDir, "invalid.doc", "Not a MIME file")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertDirectory(tmpDir, true, false)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertDirectory with verbose failed: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify skip message for non-Confluence file
	if !strings.Contains(output, "Skipping (not Confluence MIME)") {
		t.Errorf("Expected 'Skipping (not Confluence MIME)' message in verbose output, got: %s", output)
	}
}

func TestConvertDirectory_EmptyMessages(t *testing.T) {
	tmpDir := t.TempDir()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertDirectory(tmpDir, false, false)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertDirectory on empty dir failed: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify empty directory message
	if !strings.Contains(output, "No .doc files found") {
		t.Errorf("Expected 'No .doc files found' message, got: %s", output)
	}
}

func TestConvertDirectory_NoConfluenceMessages(t *testing.T) {
	tmpDir := t.TempDir()
	createPlainTextFile(t, tmpDir, "plain.doc", "Not MIME content")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := convertDirectory(tmpDir, false, false)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("convertDirectory failed: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify no Confluence files message
	if !strings.Contains(output, "No Confluence MIME exports found") {
		t.Errorf("Expected 'No Confluence MIME exports found' message, got: %s", output)
	}
}

func TestGenerateOutputPath_MoreEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "consecutive plus signs",
			input:    "my++file.doc",
			expected: "my--file.md",
		},
		{
			name:     "plus at start",
			input:    "+file.doc",
			expected: "-file.md",
		},
		{
			name:     "plus at end before extension",
			input:    "file+.doc",
			expected: "file-.md",
		},
		{
			name:     "hidden file",
			input:    ".hidden.doc",
			expected: ".hidden.md",
		},
		{
			name:     "deeply nested path",
			input:    "/a/b/c/d/e/file.doc",
			expected: "/a/b/c/d/e/file.md",
		},
		{
			name:     "current directory path",
			input:    "./file.doc",
			expected: "file.md",
		},
		{
			name:     "parent directory path",
			input:    "../file.doc",
			expected: "../file.md",
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
