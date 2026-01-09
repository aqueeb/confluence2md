package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"confluence2md/converter"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	repoURL = "https://github.com/aqueeb/confluence2md"
)

func main() {
	// Define flags
	outputPath := flag.String("o", "", "Output file path (default: input with .md extension)")
	outputLong := flag.String("output", "", "Output file path (default: input with .md extension)")
	dirMode := flag.String("dir", "", "Convert all .doc files in directory")
	verbose := flag.Bool("v", false, "Verbose output")
	verboseLong := flag.Bool("verbose", false, "Verbose output")
	dryRun := flag.Bool("dry-run", false, "Show what would be converted without writing")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "confluence2md - Convert Confluence MIME exports to Markdown\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags] <input.doc>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dir <directory>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s document.doc                    Convert single file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s document.doc -o output.md       Convert with custom output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dir ./docs                    Convert all .doc files in directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dir ./docs --dry-run          Preview conversions\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("confluence2md %s\n", version)
		if commit != "none" {
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		}
		os.Exit(0)
	}

	// Merge short and long flag variants
	if *outputLong != "" && *outputPath == "" {
		outputPath = outputLong
	}
	if *verboseLong && !*verbose {
		verbose = verboseLong
	}

	// Check pandoc availability
	if err := converter.CheckPandoc(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Directory mode
	if *dirMode != "" {
		if err := convertDirectory(*dirMode, *verbose, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !*dryRun {
			printStarPrompt()
		}
		return
	}

	// Single file mode
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := args[0]
	output := *outputPath
	if output == "" {
		output = generateOutputPath(inputPath)
	}

	if err := convertFile(inputPath, output, *verbose, *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !*dryRun {
		printStarPrompt()
	}
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
		if converter.IsConfluenceMIME(match) {
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
	if !converter.IsConfluenceMIME(inputPath) {
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
