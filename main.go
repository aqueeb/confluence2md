// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aqueeb/confluence2md/converter"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	repoURL = "https://github.com/aqueeb/confluence2md"
)

// config holds the parsed command-line configuration
type config struct {
	outputPath  string
	dirMode     string
	verbose     bool
	dryRun      bool
	showVersion bool
	args        []string
}

// parseFlags parses command-line flags and returns a config.
// Uses the provided FlagSet to allow testing without affecting global state.
func parseFlags(args []string, output io.Writer) (*config, error) {
	fs := flag.NewFlagSet("confluence2md", flag.ContinueOnError)
	fs.SetOutput(output)

	outputPath := fs.String("o", "", "Output file path (default: input with .md extension)")
	outputLong := fs.String("output", "", "Output file path (default: input with .md extension)")
	dirMode := fs.String("dir", "", "Convert all .doc files in directory")
	verbose := fs.Bool("v", false, "Verbose output")
	verboseLong := fs.Bool("verbose", false, "Verbose output")
	dryRun := fs.Bool("dry-run", false, "Show what would be converted without writing")
	showVersion := fs.Bool("version", false, "Show version")

	fs.Usage = func() {
		fmt.Fprintf(output, "confluence2md - Convert Confluence MIME exports to Markdown\n\n")
		fmt.Fprintf(output, "Usage:\n")
		fmt.Fprintf(output, "  confluence2md [flags] <input.doc>\n")
		fmt.Fprintf(output, "  confluence2md --dir <directory>\n\n")
		fmt.Fprintf(output, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(output, "\nExamples:\n")
		fmt.Fprintf(output, "  confluence2md document.doc                    Convert single file\n")
		fmt.Fprintf(output, "  confluence2md document.doc -o output.md       Convert with custom output\n")
		fmt.Fprintf(output, "  confluence2md --dir ./docs                    Convert all .doc files in directory\n")
		fmt.Fprintf(output, "  confluence2md --dir ./docs --dry-run          Preview conversions\n")
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Merge short and long flag variants
	outPath := *outputPath
	if *outputLong != "" && outPath == "" {
		outPath = *outputLong
	}
	isVerbose := *verbose || *verboseLong

	return &config{
		outputPath:  outPath,
		dirMode:     *dirMode,
		verbose:     isVerbose,
		dryRun:      *dryRun,
		showVersion: *showVersion,
		args:        fs.Args(),
	}, nil
}

// run executes the main logic and returns an exit code.
// This function is testable as it doesn't call os.Exit directly.
func run(cfg *config) int {
	// Handle version flag
	if cfg.showVersion {
		fmt.Printf("confluence2md %s\n", version)
		if commit != "none" {
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		}
		return 0
	}

	// Check pandoc availability
	if err := converter.CheckPandoc(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Directory mode
	if cfg.dirMode != "" {
		if err := convertDirectory(cfg.dirMode, cfg.verbose, cfg.dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !cfg.dryRun {
			printStarPrompt()
		}
		return 0
	}

	// Single file mode
	if len(cfg.args) < 1 {
		fmt.Fprintf(os.Stderr, "confluence2md - Convert Confluence MIME exports to Markdown\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  confluence2md [flags] <input.doc>\n")
		fmt.Fprintf(os.Stderr, "  confluence2md --dir <directory>\n\n")
		fmt.Fprintf(os.Stderr, "Run 'confluence2md --help' for more information.\n")
		return 1
	}

	inputPath := cfg.args[0]
	output := cfg.outputPath
	if output == "" {
		output = generateOutputPath(inputPath)
	}

	if err := convertFile(inputPath, output, cfg.verbose, cfg.dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if !cfg.dryRun {
		printStarPrompt()
	}
	return 0
}

func main() {
	cfg, err := parseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(run(cfg))
}

// convertDirectory converts all .doc files in a directory.
func convertDirectory(dir string, verbose, dryRun bool) error {
	pattern := filepath.Join(dir, "*.doc")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob directory: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No .doc files found in directory")
		return nil
	}

	// Filter to only Confluence MIME files
	var confluenceFiles []string
	for _, match := range matches {
		isConfluence, err := converter.IsConfluenceMIME(match)
		if err != nil {
			if verbose {
				fmt.Printf("Skipping (error reading file): %s: %v\n", match, err)
			}
			continue
		}
		if isConfluence {
			confluenceFiles = append(confluenceFiles, match)
		} else if verbose {
			fmt.Printf("Skipping (not Confluence MIME): %s\n", match)
		}
	}

	if len(confluenceFiles) == 0 {
		fmt.Println("No Confluence MIME exports found in directory")
		return nil
	}

	fmt.Printf("Found %d Confluence export(s) to convert\n", len(confluenceFiles))

	successCount := 0
	for _, inputPath := range confluenceFiles {
		outputPath := generateOutputPath(inputPath)
		if err := convertFile(inputPath, outputPath, verbose, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to convert %s: %v\n", inputPath, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\nConverted %d/%d files\n", successCount, len(confluenceFiles))
	return nil
}

// convertFile converts a single file.
func convertFile(inputPath, outputPath string, verbose, dryRun bool) error {
	if verbose {
		fmt.Printf("Converting: %s -> %s\n", inputPath, outputPath)
	}

	if dryRun {
		fmt.Printf("[dry-run] Would convert: %s -> %s\n", inputPath, outputPath)
		return nil
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Verify it's a Confluence MIME export
	isConfluence, err := converter.IsConfluenceMIME(inputPath)
	if err != nil {
		return fmt.Errorf("failed to check file format: %w", err)
	}
	if !isConfluence {
		return fmt.Errorf("file does not appear to be a Confluence MIME export: %s", inputPath)
	}

	// Extract HTML from MIME
	if verbose {
		fmt.Println("  Extracting HTML from MIME...")
	}
	html, err := converter.ExtractHTMLFromMIME(inputPath)
	if err != nil {
		return fmt.Errorf("failed to extract HTML: %w", err)
	}

	// Convert to Markdown
	if verbose {
		fmt.Println("  Converting HTML to Markdown...")
	}
	markdown, err := converter.ConvertHTMLToMarkdown(html)
	if err != nil {
		return fmt.Errorf("failed to convert to Markdown: %w", err)
	}

	// Write output
	if verbose {
		fmt.Println("  Writing output...")
	}
	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if !verbose {
		fmt.Printf("Converted: %s -> %s\n", filepath.Base(inputPath), filepath.Base(outputPath))
	} else {
		fmt.Printf("  Done: %s\n", outputPath)
	}

	return nil
}

// generateOutputPath creates the output path from an input path.
// Replaces .doc with .md and converts + to - in filename.
func generateOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)

	// Remove .doc extension
	name := strings.TrimSuffix(base, ".doc")

	// Replace + with - for cleaner filenames
	name = strings.ReplaceAll(name, "+", "-")

	// Add .md extension
	return filepath.Join(dir, name+".md")
}

// printStarPrompt prints a message asking users to star the repo and support.
func printStarPrompt() {
	fmt.Println()
	fmt.Println("Glad I could help! If you found this useful:")
	fmt.Printf("   ⭐ Star the repo: %s\n", repoURL)
	fmt.Println("   ☕ Buy me a coffee: https://buymeacoffee.com/aqueeb")
}
