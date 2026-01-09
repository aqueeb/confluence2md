// Package pandoc provides embedded Pandoc binary support with automatic extraction.
package pandoc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Version is the embedded Pandoc version
const Version = "3.6.4"

var (
	extractOnce   sync.Once
	extractedPath string
	extractErr    error
)

// EnsureExtracted extracts the embedded Pandoc binary to a cache location
// and returns the path. Safe for concurrent use. Subsequent calls return
// the cached path without re-extraction.
func EnsureExtracted() (string, error) {
	extractOnce.Do(func() {
		extractedPath, extractErr = extractBinary()
	})
	return extractedPath, extractErr
}

// extractBinary extracts the embedded binary to a persistent cache location.
func extractBinary() (string, error) {
	// Get user cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to temp directory
		cacheDir = os.TempDir()
	}

	// Create versioned cache directory
	pandocDir := filepath.Join(cacheDir, "confluence2md", fmt.Sprintf("pandoc-%s", Version))
	if err := os.MkdirAll(pandocDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Determine binary name
	binaryName := getBinaryName()
	binaryPath := filepath.Join(pandocDir, binaryName)

	// Check if binary already exists and has correct size
	if info, err := os.Stat(binaryPath); err == nil {
		expectedSize := int64(len(embeddedBinary))
		if info.Size() == expectedSize {
			// Binary exists and matches expected size, verify it's executable
			if err := verifyExecutable(binaryPath); err == nil {
				return binaryPath, nil
			}
		}
		// Size mismatch or not executable, remove and re-extract
		os.Remove(binaryPath)
	}

	// Write embedded binary
	if err := os.WriteFile(binaryPath, embeddedBinary, 0755); err != nil {
		return "", fmt.Errorf("failed to write pandoc binary: %w", err)
	}

	// Verify extraction
	if err := verifyExecutable(binaryPath); err != nil {
		os.Remove(binaryPath)
		return "", fmt.Errorf("extracted binary verification failed: %w", err)
	}

	return binaryPath, nil
}

// getBinaryName returns the platform-specific binary name.
func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "pandoc.exe"
	}
	return "pandoc"
}

// verifyExecutable checks if the binary is executable by running --version.
func verifyExecutable(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*1000*1000*1000) // 10 seconds
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary not executable: %w", err)
	}

	if !strings.Contains(string(output), "pandoc") {
		return fmt.Errorf("unexpected output from pandoc --version")
	}

	return nil
}

// Run executes pandoc with the given arguments and returns combined output.
func Run(ctx context.Context, args ...string) ([]byte, error) {
	pandocPath, err := EnsureExtracted()
	if err != nil {
		return nil, fmt.Errorf("failed to extract pandoc: %w", err)
	}

	cmd := exec.CommandContext(ctx, pandocPath, args...)
	return cmd.CombinedOutput()
}

// Convert performs a pandoc conversion with input from stdin.
func Convert(ctx context.Context, input []byte, from, to string, extraArgs ...string) ([]byte, error) {
	pandocPath, err := EnsureExtracted()
	if err != nil {
		return nil, fmt.Errorf("failed to extract pandoc: %w", err)
	}

	args := []string{"-f", from, "-t", to}
	args = append(args, extraArgs...)

	cmd := exec.CommandContext(ctx, pandocPath, args...)
	cmd.Stdin = bytes.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pandoc error: %w: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// GetVersion returns the version string from the embedded pandoc.
func GetVersion(ctx context.Context) (string, error) {
	output, err := Run(ctx, "--version")
	if err != nil {
		return "", err
	}

	// Parse first line for version
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", fmt.Errorf("failed to parse pandoc version")
}

// GetPath returns the path to the extracted pandoc binary.
// Returns empty string if not yet extracted.
func GetPath() string {
	return extractedPath
}

// Cleanup removes the extracted binary and its directory.
// This is optional - the binary is cached for reuse across runs.
func Cleanup() error {
	if extractedPath == "" {
		return nil
	}

	dir := filepath.Dir(extractedPath)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to cleanup pandoc: %w", err)
	}

	// Reset state so next call will re-extract
	extractedPath = ""
	extractOnce = sync.Once{}

	return nil
}

// IsEmbedded returns true if a pandoc binary is embedded in this build.
func IsEmbedded() bool {
	return len(embeddedBinary) > 0
}

// EmbeddedSize returns the size of the embedded binary in bytes.
func EmbeddedSize() int {
	return len(embeddedBinary)
}
