package places

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ScraperFunc runs the scraper with the given options and returns parsed results.
// In production: writes temp input file, shells out to binary, reads JSON output.
// In tests: returns canned data without any I/O.
type ScraperFunc func(ctx context.Context, opts ScraperOptions) (ScraperResult, error)

// ScraperOptions holds all configuration for a scraper invocation.
type ScraperOptions struct {
	Queries     []string // one query per line (or Google Maps URLs for lookup)
	Geo         string   // "lat,lng" for geo-targeting
	Zoom        int      // Google Maps zoom level (affects search area, 1-21)
	Depth       int      // pagination depth (1=first page only)
	Email       bool     // extract emails from business websites
	FastMode    bool     // HTTP-only mode, no browser
	Concurrency int      // number of concurrent scrapers
	Lang        string   // language code
	Limit       int      // post-processing limit on results
}

// ScraperResult holds the parsed output from one scraper invocation.
type ScraperResult struct {
	Entries []Entry
}

// defaultScraperBinary resolves the path to the google-maps-scraper binary.
func defaultScraperBinary() string {
	if bin := os.Getenv("GOOGLE_MAPS_SCRAPER_BIN"); bin != "" {
		return bin
	}
	if path, err := exec.LookPath("google-maps-scraper"); err == nil {
		return path
	}
	return "google-maps-scraper"
}

// defaultScraperFunc returns a ScraperFunc that shells out to the given binary.
func defaultScraperFunc(binary string) ScraperFunc {
	return func(ctx context.Context, opts ScraperOptions) (ScraperResult, error) {
		tmpDir, err := os.MkdirTemp("", "places-scraper-*")
		if err != nil {
			return ScraperResult{}, fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		inputFile := filepath.Join(tmpDir, "input.txt")
		outputFile := filepath.Join(tmpDir, "output.json")

		// Write queries (one per line)
		content := strings.Join(opts.Queries, "\n") + "\n"
		if err := os.WriteFile(inputFile, []byte(content), 0o600); err != nil {
			return ScraperResult{}, fmt.Errorf("write input: %w", err)
		}

		// Build command args
		args := []string{
			"-input", inputFile,
			"-results", outputFile,
			"-json",
			"-exit-on-inactivity", "3m",
		}
		if opts.Geo != "" {
			args = append(args, "-geo", opts.Geo)
		}
		if opts.Zoom > 0 {
			args = append(args, "-zoom", strconv.Itoa(opts.Zoom))
		}
		if opts.Depth > 0 {
			args = append(args, "-depth", strconv.Itoa(opts.Depth))
		}
		if opts.Email {
			args = append(args, "-email")
		}
		if opts.Concurrency > 0 {
			args = append(args, "-c", strconv.Itoa(opts.Concurrency))
		}
		if opts.Lang != "" {
			args = append(args, "-lang", opts.Lang)
		}

		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return ScraperResult{}, fmt.Errorf("run scraper: %w", err)
		}

		entries, err := parseOutputFile(outputFile)
		if err != nil {
			return ScraperResult{}, err
		}

		// Apply post-processing limit
		if opts.Limit > 0 && len(entries) > opts.Limit {
			entries = entries[:opts.Limit]
		}

		return ScraperResult{Entries: entries}, nil
	}
}

// parseOutputFile reads the scraper output file. Handles both JSON array and
// NDJSON (newline-delimited JSON) formats.
func parseOutputFile(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	data = trimBOM(data)
	if len(data) == 0 {
		return nil, nil
	}

	// Try JSON array first
	if data[0] == '[' {
		var entries []Entry
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("parse JSON array: %w", err)
		}
		return entries, nil
	}

	// Fall back to NDJSON (one JSON object per line)
	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB line buffer
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse NDJSON line: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, scanner.Err()
}

// trimBOM removes a UTF-8 BOM if present.
func trimBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
		return data[3:]
	}
	return data
}
