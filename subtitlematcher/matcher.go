// Package subtitlematcher provides functionality to match and rename subtitle files
// to correspond with their associated video files.
//
// The main type VideoSubtitleMatcher uses intelligent matching algorithms to pair
// subtitle files with video files based on filename similarity, even when the
// naming conventions differ (such as YouTube downloads with different patterns).
package subtitlematcher

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// VideoSubtitleMatcher handles matching and renaming subtitle files to match video files.
// It supports various video and subtitle formats and uses configurable similarity
// algorithms to ensure accurate matching.
type VideoSubtitleMatcher struct {
	videoExtensions     []string  // Supported video file extensions
	subtitleExtensions  []string  // Supported subtitle file extensions
	directory           string    // Working directory
	similarityThreshold float64   // Minimum similarity score for matching (0.0-1.0)
	recursive           bool      // Whether to scan directories recursively
	dryRun              bool      // Whether to perform actual file operations
	verbose             bool      // Whether to output detailed information
	ignoreExisting      bool      // Whether to skip files that are already correctly named
}

// Option defines a functional option for configuring VideoSubtitleMatcher.
type Option func(*VideoSubtitleMatcher)

// VideoExtensions sets custom video file extensions.
// Default: [".mkv", ".mp4", ".avi", ".mov", ".webm"]
func VideoExtensions(extensions []string) Option {
	return func(vsm *VideoSubtitleMatcher) {
		vsm.videoExtensions = extensions
	}
}

// SubtitleExtensions sets custom subtitle file extensions.
// Default: [".srt", ".ass", ".vtt"]
func SubtitleExtensions(extensions []string) Option {
	return func(vsm *VideoSubtitleMatcher) {
		vsm.subtitleExtensions = extensions
	}
}

// SimilarityThreshold sets the minimum similarity threshold for matching.
// Values range from 0.0 (no similarity required) to 1.0 (exact match required).
// Default: 0.6
func SimilarityThreshold(threshold float64) Option {
	return func(vsm *VideoSubtitleMatcher) {
		if threshold >= 0.0 && threshold <= 1.0 {
			vsm.similarityThreshold = threshold
		}
	}
}

// Recursive enables or disables recursive directory scanning.
// Default: true
func Recursive(recursive bool) Option {
	return func(vsm *VideoSubtitleMatcher) {
		vsm.recursive = recursive
	}
}

// DryRun enables or disables dry run mode.
// In dry run mode, no actual file operations are performed.
// Default: true
func DryRun(dryRun bool) Option {
	return func(vsm *VideoSubtitleMatcher) {
		vsm.dryRun = dryRun
	}
}

// Verbose enables or disables verbose output.
// Default: true
func Verbose(verbose bool) Option {
	return func(vsm *VideoSubtitleMatcher) {
		vsm.verbose = verbose
	}
}

// IgnoreExisting sets whether to ignore already correctly named files.
// Default: false
func IgnoreExisting(ignore bool) Option {
	return func(vsm *VideoSubtitleMatcher) {
		vsm.ignoreExisting = ignore
	}
}

// New creates a new VideoSubtitleMatcher instance with the specified directory
// and optional configuration options.
//
// The directory parameter specifies the root directory to scan for video and subtitle files.
// Additional options can be provided to customize the matching behavior.
//
// Example:
//   matcher := subtitlematcher.New("/path/to/videos",
//       subtitlematcher.SimilarityThreshold(0.8),
//       subtitlematcher.DryRun(false),
//   )
func New(directory string, options ...Option) *VideoSubtitleMatcher {
	// Initialize with sensible defaults
	vsm := &VideoSubtitleMatcher{
		videoExtensions:     []string{".mkv", ".mp4", ".avi", ".mov", ".webm"},
		subtitleExtensions:  []string{".srt", ".ass", ".vtt"},
		directory:           directory,
		similarityThreshold: 0.6,
		recursive:           true,
		dryRun:              true,
		verbose:             true,
		ignoreExisting:      false,
	}

	// Apply functional options
	for _, option := range options {
		option(vsm)
	}

	return vsm
}

// scanFiles scans the configured directory and returns lists of video and subtitle files.
// The scanning behavior (recursive vs non-recursive) is controlled by the recursive option.
func (vsm *VideoSubtitleMatcher) scanFiles() ([]string, []string, error) {
	var videoFiles, subtitleFiles []string

	if vsm.recursive {
		err := filepath.Walk(vsm.directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))

			for _, videoExt := range vsm.videoExtensions {
				if ext == videoExt {
					videoFiles = append(videoFiles, path)
					return nil
				}
			}

			for _, subtitleExt := range vsm.subtitleExtensions {
				if ext == subtitleExt {
					subtitleFiles = append(subtitleFiles, path)
					return nil
				}
			}

			return nil
		})
		return videoFiles, subtitleFiles, err
	} else {
		// Non-recursive scan - only current directory
		entries, err := os.ReadDir(vsm.directory)
		if err != nil {
			return nil, nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(vsm.directory, entry.Name())
			ext := strings.ToLower(filepath.Ext(entry.Name()))

			for _, videoExt := range vsm.videoExtensions {
				if ext == videoExt {
					videoFiles = append(videoFiles, fullPath)
					break
				}
			}

			for _, subtitleExt := range vsm.subtitleExtensions {
				if ext == subtitleExt {
					subtitleFiles = append(subtitleFiles, fullPath)
					break
				}
			}
		}
		return videoFiles, subtitleFiles, nil
	}
}

// normalizeTitle normalizes video/subtitle titles for comparison by removing
// platform-specific patterns and standardizing the format.
//
// This function handles common patterns like:
// - YouTube IDs in brackets: [ABC123]
// - YouTube subtitle suffixes: -_YouTube-zh-CN-dual-double
// - Underscores to spaces conversion
// - Character normalization (e.g., ？ to ?)
func (vsm *VideoSubtitleMatcher) normalizeTitle(title string) string {
	// Remove YouTube ID pattern [xxxxx] from video files
	re := regexp.MustCompile(`\[[A-Za-z0-9_-]+\]`)
	title = re.ReplaceAllString(title, "")

	// Remove YouTube subtitle patterns
	title = strings.ReplaceAll(title, "-_YouTube-zh-CN-dual-double", "")
	title = strings.ReplaceAll(title, "_-_YouTube", "")

	// Replace underscores with spaces and normalize
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "？", "?")

	// Remove extra spaces and convert to lowercase
	title = strings.TrimSpace(title)
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return strings.ToLower(title)
}

// findBestMatch finds the best matching video file for a given subtitle file
// using fuzzy string matching based on the longest common subsequence algorithm.
//
// Returns the path of the best matching video file and the similarity score (0.0-1.0).
func (vsm *VideoSubtitleMatcher) findBestMatch(subtitlePath string, videoFiles []string) (string, float64) {
	subtitleName := strings.TrimSuffix(filepath.Base(subtitlePath), filepath.Ext(subtitlePath))
	normalizedSubtitle := vsm.normalizeTitle(subtitleName)

	var bestMatch string
	var bestScore float64

	for _, videoPath := range videoFiles {
		videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
		normalizedVideo := vsm.normalizeTitle(videoName)

		score := vsm.calculateSimilarity(normalizedSubtitle, normalizedVideo)
		if score > bestScore {
			bestScore = score
			bestMatch = videoPath
		}
	}

	return bestMatch, bestScore
}

// calculateSimilarity calculates the similarity between two strings using the
// longest common subsequence (LCS) algorithm.
//
// Returns a score between 0.0 (no similarity) and 1.0 (identical).
func (vsm *VideoSubtitleMatcher) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	lcs := vsm.longestCommonSubsequence(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 0.0
	}

	return float64(lcs) / float64(maxLen)
}

// longestCommonSubsequence calculates the length of the longest common subsequence
// between two strings using dynamic programming.
func (vsm *VideoSubtitleMatcher) longestCommonSubsequence(s1, s2 string) int {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	return dp[m][n]
}

// MatchResult represents the result of a subtitle matching operation.
type MatchResult struct {
	SubtitlePath    string  // Original subtitle file path
	VideoPath       string  // Matched video file path
	NewSubtitlePath string  // New subtitle file path after renaming
	Similarity      float64 // Similarity score (0.0-1.0)
	Renamed         bool    // Whether the file was actually renamed
	Error           error   // Any error that occurred during renaming
}

// Match performs the subtitle matching and renaming operation.
// Returns a slice of MatchResult containing details about each processed subtitle file.
//
// This is the main entry point for the subtitle matching functionality.
func (vsm *VideoSubtitleMatcher) Match() ([]MatchResult, error) {
	videoFiles, subtitleFiles, err := vsm.scanFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	if vsm.verbose {
		fmt.Printf("Found %d video files and %d subtitle files\n", len(videoFiles), len(subtitleFiles))
	}

	var results []MatchResult

	for _, subtitlePath := range subtitleFiles {
		bestMatch, score := vsm.findBestMatch(subtitlePath, videoFiles)

		result := MatchResult{
			SubtitlePath: subtitlePath,
			VideoPath:    bestMatch,
			Similarity:   score,
		}

		if score >= vsm.similarityThreshold {
			videoBaseName := strings.TrimSuffix(filepath.Base(bestMatch), filepath.Ext(bestMatch))
			subtitleExt := filepath.Ext(subtitlePath)
			newSubtitlePath := filepath.Join(filepath.Dir(subtitlePath), videoBaseName+subtitleExt)
			result.NewSubtitlePath = newSubtitlePath

			// Skip if already correctly named and ignoreExisting is true
			if vsm.ignoreExisting && subtitlePath == newSubtitlePath {
				continue
			}

			if vsm.verbose {
				fmt.Printf("\nMatch found (%.2f similarity):\n", score)
				fmt.Printf("  Subtitle: %s\n", filepath.Base(subtitlePath))
				fmt.Printf("  Video:    %s\n", filepath.Base(bestMatch))
				fmt.Printf("  New name: %s\n", filepath.Base(newSubtitlePath))
			}

			if !vsm.dryRun {
				if subtitlePath != newSubtitlePath {
					err := os.Rename(subtitlePath, newSubtitlePath)
					if err != nil {
						result.Error = err
						if vsm.verbose {
							fmt.Printf("  Error renaming: %v\n", err)
						}
					} else {
						result.Renamed = true
						if vsm.verbose {
							fmt.Printf("  ✓ Renamed successfully\n")
						}
					}
				} else {
					result.Renamed = true
					if vsm.verbose {
						fmt.Printf("  ✓ Already correctly named\n")
					}
				}
			}
		} else {
			if vsm.verbose {
				fmt.Printf("\nNo good match found for: %s (best score: %.2f)\n", filepath.Base(subtitlePath), score)
			}
		}

		results = append(results, result)
	}

	if vsm.verbose {
		matchCount := 0
		for _, result := range results {
			if result.Similarity >= vsm.similarityThreshold {
				matchCount++
			}
		}

		if vsm.dryRun {
			fmt.Printf("\nDry run completed. %d subtitles would be renamed.\n", matchCount)
			fmt.Println("Use DryRun(false) option to perform actual renaming.")
		} else {
			fmt.Printf("\nRenaming completed. %d subtitles processed.\n", matchCount)
		}
	}

	return results, nil
}