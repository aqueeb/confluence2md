package converter

import (
	"os"
	"path/filepath"
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
	if !IsConfluenceMIME(validFile) {
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
	if IsConfluenceMIME(invalidFile) {
		t.Error("Expected invalid file to return false")
	}

	// Test non-existent file
	if IsConfluenceMIME("/nonexistent/file.doc") {
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
	if !contains(html, "<h1>Hello World</h1>") {
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
	if !contains(html, "Test = Value") {
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
