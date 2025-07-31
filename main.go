package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/krmmzs/subtitle-matcher/subtitlematcher"
)

// Config holds the command line configuration
type Config struct {
	Directory   string
	ExecuteMode bool
}

// parseArgs parses command line arguments and returns configuration
func parseArgs() Config {
	var config Config

	if len(os.Args) < 2 {
		config.Directory = "."
	} else {
		config.Directory = os.Args[1]
	}

	config.ExecuteMode = false
	for _, arg := range os.Args {
		if arg == "-execute" || arg == "--execute" {
			config.ExecuteMode = true
			break
		}
	}

	return config
}

// validateDirectory checks if the directory exists
func validateDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directory)
	}
	return nil
}

// runBasicExample demonstrates basic usage with default settings
func runBasicExample(directory string) error {
	fmt.Println("=== Example 1: Basic usage (dry run) ===")
	matcher := subtitlematcher.New(directory)
	results, err := matcher.Match()
	if err != nil {
		return fmt.Errorf("error in basic example: %w", err)
	}

	fmt.Printf("Processed %d subtitle files\n", len(results))
	return nil
}

// runHighThresholdExample demonstrates usage with high similarity threshold
func runHighThresholdExample(directory string, executeMode bool) error {
	fmt.Println("\n=== Example 2: Execute with high similarity threshold ===")
	matcher := subtitlematcher.New(directory,
		subtitlematcher.DryRun(!executeMode),
		subtitlematcher.SimilarityThreshold(0.8),
		subtitlematcher.Verbose(true),
	)

	results, err := matcher.Match()
	if err != nil {
		return fmt.Errorf("error in high threshold example: %w", err)
	}

	successCount := countSuccessfulRenames(results)
	fmt.Printf("Successfully processed %d subtitle files\n", successCount)
	return nil
}

// countSuccessfulRenames counts how many files were successfully renamed
func countSuccessfulRenames(results []subtitlematcher.MatchResult) int {
	count := 0
	for _, result := range results {
		if result.Renamed && result.Error == nil {
			count++
		}
	}
	return count
}

// runCustomConfigExample demonstrates custom configuration
func runCustomConfigExample(directory string, executeMode bool) error {
	fmt.Println("\n=== Example 3: Custom configuration ===")
	matcher := subtitlematcher.New(directory,
		subtitlematcher.VideoExtensions([]string{".mkv", ".mp4", ".webm"}),
		subtitlematcher.SubtitleExtensions([]string{".srt"}),
		subtitlematcher.SimilarityThreshold(0.7),
		subtitlematcher.Recursive(true),
		subtitlematcher.DryRun(!executeMode),
		subtitlematcher.Verbose(false), // Quiet mode for this example
		subtitlematcher.IgnoreExisting(true),
	)

	results, err := matcher.Match()
	if err != nil {
		return fmt.Errorf("error in custom config example: %w", err)
	}

	displayDetailedResults(results, executeMode)
	return nil
}

// displayDetailedResults shows detailed results for the custom config example
func displayDetailedResults(results []subtitlematcher.MatchResult, executeMode bool) {
	if len(results) == 0 {
		return
	}

	fmt.Println("Detailed results:")
	for _, result := range results {
		if result.Similarity >= 0.7 {
			status := determineStatus(result, executeMode)
			subtitleName := extractFileName(result.SubtitlePath, ".srt")
			newSubtitleName := extractFileName(result.NewSubtitlePath, ".srt")

			fmt.Printf("  %s (%.2f similarity) -> %s [%s]\n",
				subtitleName, result.Similarity, newSubtitleName, status)
		}
	}
}

// determineStatus determines the status string for a match result
func determineStatus(result subtitlematcher.MatchResult, executeMode bool) string {
	if !executeMode {
		return "WOULD RENAME"
	}
	if result.Renamed && result.Error == nil {
		return "RENAMED"
	}
	if result.Error != nil {
		return fmt.Sprintf("ERROR: %v", result.Error)
	}
	return "NO CHANGE"
}

// extractFileName extracts the filename without path and extension
func extractFileName(fullPath, extension string) string {
	lastIndex := strings.LastIndex(fullPath, "/")
	if lastIndex == -1 {
		lastIndex = 0
	} else {
		lastIndex++
	}
	return strings.TrimSuffix(fullPath[lastIndex:], extension)
}

// printUsageInformation displays usage information to the user
func printUsageInformation(executeMode bool) {
	if !executeMode {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("All examples ran in dry-run mode.")
		fmt.Println("Add -execute flag to perform actual renaming.")
		printUsageExamples()
	} else {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("File renaming operations completed.")
	}
}

// printUsageExamples prints command line usage examples
func printUsageExamples() {
	fmt.Println("\nUsage:")
	fmt.Println("  go run main.go [directory] [-execute]")
	fmt.Println("\nExamples:")
	fmt.Println("  go run main.go                    # Dry run in current directory")
	fmt.Println("  go run main.go /path/to/videos    # Dry run in specified directory")
	fmt.Println("  go run main.go . -execute         # Execute renaming in current directory")
}

func main() {
	config := parseArgs()

	// Validate directory exists
	if err := validateDirectory(config.Directory); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Run examples
	if err := runBasicExample(config.Directory); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := runHighThresholdExample(config.Directory, config.ExecuteMode); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := runCustomConfigExample(config.Directory, config.ExecuteMode); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print usage information
	printUsageInformation(config.ExecuteMode)
}