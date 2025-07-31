package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/krmmzs/subtitle-matcher/subtitlematcher"
)

func main() {
	var directory string
	var executeMode bool

	// Parse command line arguments
	if len(os.Args) < 2 {
		directory = "."
	} else {
		directory = os.Args[1]
	}

	executeMode = false
	for _, arg := range os.Args {
		if arg == "-execute" || arg == "--execute" {
			executeMode = true
			break
		}
	}

	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %s\n", directory)
		os.Exit(1)
	}

	// Example 1: Basic usage (dry run by default)
	fmt.Println("=== Example 1: Basic usage (dry run) ===")
	matcher1 := subtitlematcher.New(directory)
	results1, err := matcher1.Match()
	if err != nil {
		fmt.Printf("Error in Example 1: %v\n", err)
	} else {
		fmt.Printf("Processed %d subtitle files\n", len(results1))
	}

	// Example 2: High similarity threshold with execute option
	fmt.Println("\n=== Example 2: Execute with high similarity threshold ===")
	matcher2 := subtitlematcher.New(directory,
		subtitlematcher.DryRun(!executeMode),
		subtitlematcher.SimilarityThreshold(0.8),
		subtitlematcher.Verbose(true),
	)
	results2, err := matcher2.Match()
	if err != nil {
		fmt.Printf("Error in Example 2: %v\n", err)
	} else {
		successCount := 0
		for _, result := range results2 {
			if result.Renamed && result.Error == nil {
				successCount++
			}
		}
		fmt.Printf("Successfully processed %d subtitle files\n", successCount)
	}

	// Example 3: Custom configuration with detailed results
	fmt.Println("\n=== Example 3: Custom configuration ===")
	matcher3 := subtitlematcher.New(directory,
		subtitlematcher.VideoExtensions([]string{".mkv", ".mp4", ".webm"}),
		subtitlematcher.SubtitleExtensions([]string{".srt"}),
		subtitlematcher.SimilarityThreshold(0.7),
		subtitlematcher.Recursive(true),
		subtitlematcher.DryRun(!executeMode),
		subtitlematcher.Verbose(false), // Quiet mode for this example
		subtitlematcher.IgnoreExisting(true),
	)

	results3, err := matcher3.Match()
	if err != nil {
		fmt.Printf("Error in Example 3: %v\n", err)
		os.Exit(1)
	}

	// Display detailed results for Example 3
	if len(results3) > 0 {
		fmt.Println("Detailed results:")
		for _, result := range results3 {
			if result.Similarity >= 0.7 {
				status := "WOULD RENAME"
				if !executeMode {
					// In dry run mode
				} else if result.Renamed && result.Error == nil {
					status = "RENAMED"
				} else if result.Error != nil {
					status = fmt.Sprintf("ERROR: %v", result.Error)
				}
				
				fmt.Printf("  %s (%.2f similarity) -> %s [%s]\n",
					strings.TrimSuffix(result.SubtitlePath[strings.LastIndex(result.SubtitlePath, "/")+1:], ".srt"),
					result.Similarity,
					strings.TrimSuffix(result.NewSubtitlePath[strings.LastIndex(result.NewSubtitlePath, "/")+1:], ".srt"),
					status)
			}
		}
	}

	// Summary
	if !executeMode {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("All examples ran in dry-run mode.")
		fmt.Println("Add -execute flag to perform actual renaming.")
		fmt.Println("\nUsage:")
		fmt.Println("  go run main.go [directory] [-execute]")
		fmt.Println("\nExamples:")
		fmt.Println("  go run main.go                    # Dry run in current directory")
		fmt.Println("  go run main.go /path/to/videos    # Dry run in specified directory")
		fmt.Println("  go run main.go . -execute         # Execute renaming in current directory")
	} else {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("File renaming operations completed.")
	}
}