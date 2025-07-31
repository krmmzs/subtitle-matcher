# Subtitle Matcher

[![Go Reference](https://pkg.go.dev/badge/github.com/krmmzs/subtitle-matcher.svg)](https://pkg.go.dev/github.com/krmmzs/subtitle-matcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/krmmzs/subtitle-matcher)](https://goreportcard.com/report/github.com/krmmzs/subtitle-matcher)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An intelligent video subtitle file matcher library designed with the Functional Options pattern, supporting automatic renaming of subtitle files to match their corresponding video files.

## Installation

```bash
go get github.com/krmmzs/subtitle-matcher
```

## Project Structure

```
.
├── subtitlematcher/          # Core library package
│   └── matcher.go           # Main matching logic and API
├── main.go                  # Example/CLI program
├── go.mod                   # Go module configuration
└── README.md               # Documentation
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/krmmzs/subtitle-matcher/subtitlematcher"
)

func main() {
    // Create matcher instance
    matcher := subtitlematcher.New("/path/to/videos")
    
    // Execute matching operation
    results, err := matcher.Match()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Processed %d subtitle files\n", len(results))
}
```

### Advanced Configuration

```go
matcher := subtitlematcher.New("/path/to/videos",
    subtitlematcher.SimilarityThreshold(0.8),    // Set similarity threshold
    subtitlematcher.DryRun(false),               // Execute actual renaming
    subtitlematcher.Recursive(true),             // Recursively scan subdirectories
    subtitlematcher.Verbose(true),               // Show detailed information
    subtitlematcher.IgnoreExisting(true),        // Ignore already correctly named files
)

results, err := matcher.Match()
```

### Available Options

- `VideoExtensions([]string)` - Set video file extensions
- `SubtitleExtensions([]string)` - Set subtitle file extensions  
- `SimilarityThreshold(float64)` - Set matching similarity threshold (0.0-1.0)
- `Recursive(bool)` - Whether to scan directories recursively
- `DryRun(bool)` - Whether to run in dry-run mode
- `Verbose(bool)` - Whether to show verbose output
- `IgnoreExisting(bool)` - Whether to ignore already correctly named files

### Result Processing

```go
results, err := matcher.Match()
if err != nil {
    // Handle error
}

for _, result := range results {
    fmt.Printf("Subtitle: %s\n", result.SubtitlePath)
    fmt.Printf("Video: %s\n", result.VideoPath)
    fmt.Printf("Similarity: %.2f\n", result.Similarity)
    fmt.Printf("Renamed: %t\n", result.Renamed)
    if result.Error != nil {
        fmt.Printf("Error: %v\n", result.Error)
    }
}
```

## Command Line Tool Usage

### Basic Usage

```bash
# Dry run mode (default)
go run main.go .

# Specify directory
go run main.go /path/to/videos

# Execute actual renaming
go run main.go . -execute
```

### Output Example

```
=== Example 1: Basic usage (dry run) ===
Found 17 video files and 12 subtitle files

Match found (1.00 similarity):
  Subtitle: How_to_code_-_YouTube-zh-CN-dual-double.srt
  Video:    How_to_code_[ABC123].mkv
  New name: How_to_code_[ABC123].srt

Dry run completed. 12 subtitles would be renamed.
```

## Features

### Intelligent Matching Algorithm
- Uses Longest Common Subsequence (LCS) algorithm to calculate filename similarity
- Automatically handles different naming patterns from YouTube downloads
- Supports configurable similarity thresholds

### File Format Support
- **Video formats**: `.mkv`, `.mp4`, `.avi`, `.mov`, `.webm`
- **Subtitle formats**: `.srt`, `.ass`, `.vtt`

### Safety Features
- Default dry-run mode to preview operation results
- Detailed error handling and status reporting
- Optional ignore functionality for existing files

### Flexible Configuration
- Functional Options pattern for flexible parameter combinations
- Sensible defaults, ready to use out of the box
- Backward-compatible API design

## Algorithm Overview

The program uses the following steps for matching:

1. **File Scanning**: Recursively or non-recursively scan specified directory
2. **Title Normalization**: Remove special identifiers and standardize format
3. **Similarity Calculation**: Use LCS algorithm to calculate string similarity
4. **Best Match Selection**: Choose highest similarity match above threshold
5. **File Renaming**: Execute or simulate renaming operations based on configuration

## Use Cases

This library primarily solves the problem where video and subtitle files downloaded from platforms like YouTube have mismatched names, preventing media players from automatically loading subtitles.

**Before:**
```
How_to_code_[ABC123].mkv
How_to_code_-_YouTube-zh-CN-dual-double.srt
```

**After:**
```
How_to_code_[ABC123].mkv
How_to_code_[ABC123].srt  ← Now matches video name
```

## Development

This library demonstrates elegant application of the Functional Options pattern in Go, providing:
- Clean API design
- Flexible configuration options
- Good error handling
- Comprehensive documentation

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.