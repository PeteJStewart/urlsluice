package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"flag"

	"github.com/PeteJStewart/urlsluice/internal/extractor"
	"github.com/PeteJStewart/urlsluice/internal/wordlist"
)

// Config holds the command-line configuration
type Config struct {
	FilePath         string
	UUIDVersion      int
	ExtractEmails    bool
	ExtractDomains   bool
	ExtractIPs       bool
	ExtractParams    bool
	Silent           bool
	GenerateWordlist bool
}

func getProgramName() string {
	name := filepath.Base(os.Args[0])
	// Handle both temporary build paths and direct go run cases
	if strings.HasPrefix(name, "/tmp/go-build") || name == "main" {
		return "urlsluice"
	}
	return name
}

// Move the help text generation to a separate function
func generateHelpText(w io.Writer, progName string) {
	fmt.Fprintf(w, "URL Sluice - Extract patterns from text files\n\n")
	fmt.Fprintf(w, "Usage: %s [options]\n\n", progName)
	fmt.Fprintf(w, "Options:\n")
	fmt.Fprintf(w, "  -file string\n")
	fmt.Fprintf(w, "        Path to the input file (required)\n")
	fmt.Fprintf(w, "  -uuid int\n")
	fmt.Fprintf(w, "        UUID version to extract (1-5) (default 4)\n")
	fmt.Fprintf(w, "  -emails\n")
	fmt.Fprintf(w, "        Extract email addresses\n")
	fmt.Fprintf(w, "  -domains\n")
	fmt.Fprintf(w, "        Extract domain names\n")
	fmt.Fprintf(w, "  -ips\n")
	fmt.Fprintf(w, "        Extract IP addresses\n")
	fmt.Fprintf(w, "  -queryParams\n")
	fmt.Fprintf(w, "        Extract query parameters\n")
	fmt.Fprintf(w, "  -silent\n")
	fmt.Fprintf(w, "        Output data without titles\n")
	fmt.Fprintf(w, "  -wordlist\n")
	fmt.Fprintf(w, "        Generate a wordlist from URLs in file\n\n")
	fmt.Fprintf(w, "Examples:\n")
	fmt.Fprintf(w, "  Extract all patterns:\n")
	fmt.Fprintf(w, "    %s -file input.txt -emails -domains -ips -queryParams\n\n", progName)
	fmt.Fprintf(w, "  Extract only domains and IPs in silent mode:\n")
	fmt.Fprintf(w, "    %s -file input.txt -domains -ips -silent\n\n", progName)
	fmt.Fprintf(w, "  Extract specific UUID version:\n")
	fmt.Fprintf(w, "    %s -file input.txt -uuid 4\n", progName)
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Parse flags
	config, err := parseFlags()
	if err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	// Open and read input file
	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Handle wordlist generation
	if config.GenerateWordlist {
		urls := strings.Split(string(data), "\n")
		tokens := wordlist.GenerateWordlist(urls)
		for _, token := range tokens {
			fmt.Println(token)
		}
		return nil
	}

	// Create extractor for pattern extraction
	ext, err := extractor.New(extractor.Config{
		UUIDVersion:    config.UUIDVersion,
		ExtractEmails:  config.ExtractEmails,
		ExtractDomains: config.ExtractDomains,
		ExtractIPs:     config.ExtractIPs,
		ExtractParams:  config.ExtractParams,
	})
	if err != nil {
		return fmt.Errorf("error creating extractor: %w", err)
	}

	// Process file
	results, err := ext.Extract(ctx, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Print results
	return printResults(results, config.Silent)
}

func printResults(results extractor.Results, silent bool) error {
	printSection := func(label string, items map[string]bool) {
		if len(items) == 0 {
			return
		}

		// Convert map keys to sorted slice
		sorted := make([]string, 0, len(items))
		for item := range items {
			sorted = append(sorted, item)
		}
		sort.Strings(sorted)

		if !silent {
			fmt.Printf("\nExtracted %s:\n", label)
		}
		for _, item := range sorted {
			fmt.Println(item)
		}
	}

	printSection("UUIDs", results.UUIDs)
	printSection("Emails", results.Emails)
	printSection("Domains", results.Domains)
	printSection("IP Addresses", results.IPs)
	printSection("Query Parameters", results.Params)

	return nil
}

func parseFlags() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.FilePath, "file", "", "Path to the input file (required)")
	flag.IntVar(&config.UUIDVersion, "uuid", 4, "UUID version to extract (1-5)")
	flag.BoolVar(&config.ExtractEmails, "emails", false, "Extract email addresses")
	flag.BoolVar(&config.ExtractDomains, "domains", false, "Extract domain names")
	flag.BoolVar(&config.ExtractIPs, "ips", false, "Extract IP addresses")
	flag.BoolVar(&config.ExtractParams, "queryParams", false, "Extract query parameters")
	flag.BoolVar(&config.Silent, "silent", false, "Output data without titles")
	flag.BoolVar(&config.GenerateWordlist, "wordlist", false, "Generate a wordlist from URLs in file")

	flag.Parse()

	if config.FilePath == "" {
		return nil, fmt.Errorf("file path is required")
	}

	return config, nil
}
