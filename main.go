package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Regex patterns for UUIDs, emails, domains, IPs, and query parameters.
var (
	uuidRegexMap = map[int]*regexp.Regexp{
		1: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-1[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		2: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-2[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		3: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-3[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		4: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		5: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-5[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
	}

	emailRegex      = regexp.MustCompile(`[\w._%+-]+@[\w.-]+\.[a-zA-Z]{2,}`)
	domainRegex     = regexp.MustCompile(`https?://([a-zA-Z0-9.-]+)/?`)
	ipRegex         = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	queryParamRegex = regexp.MustCompile(`[?&]([^&=]+)=([^&=]*)`)
)

// ExtractionResults holds the extracted values.
type ExtractionResults struct {
	UUIDs   map[string]bool
	Emails  map[string]bool
	Domains map[string]bool
	IPs     map[string]bool
	Params  map[string]bool
}

// newExtractionResults creates an initialized ExtractionResults struct.
func newExtractionResults() *ExtractionResults {
	return &ExtractionResults{
		UUIDs:   make(map[string]bool),
		Emails:  make(map[string]bool),
		Domains: make(map[string]bool),
		IPs:     make(map[string]bool),
		Params:  make(map[string]bool),
	}
}

// extractData opens the file and iterates through its lines, applying the various extraction functions.
func extractData(filePath string, uuidVersion int, extractEmails, extractDomains, extractIPs, extractQueryParams, silent bool) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	results := newExtractionResults()

	// Select the UUID regex based on the provided version.
	uuidRegex, exists := uuidRegexMap[uuidVersion]
	if !exists {
		log.Fatalf("Error: Unsupported UUID version. Use 1-5.")
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Extract data from the current line.
		extractUUIDs(line, uuidRegex, results.UUIDs)
		if extractEmails {
			extractEmailsFromLine(line, results.Emails)
		}
		if extractDomains {
			extractDomainsFromLine(line, results.Domains)
		}
		if extractIPs {
			extractIPsFromLine(line, results.IPs)
		}
		if extractQueryParams {
			extractQueryParamsFromLine(line, results.Params)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Print all extracted data.
	printExtractedData("UUIDs", results.UUIDs, silent)
	printExtractedData("Email Addresses", results.Emails, silent)
	printExtractedData("Domains", results.Domains, silent)
	printExtractedData("IP Addresses", results.IPs, silent)
	printExtractedData("Query Parameters", results.Params, silent)
}

// extractUUIDs uses the given regex to find and store UUIDs.
func extractUUIDs(line string, uuidRegex *regexp.Regexp, uuidMap map[string]bool) {
	matches := uuidRegex.FindAllString(line, -1)
	for _, uuid := range matches {
		uuidMap[uuid] = true
	}
}

// extractEmailsFromLine extracts email addresses from the line.
func extractEmailsFromLine(line string, emailMap map[string]bool) {
	matches := emailRegex.FindAllString(line, -1)
	for _, email := range matches {
		emailMap[email] = true
	}
}

// extractDomainsFromLine extracts domain names from the line.
func extractDomainsFromLine(line string, domainMap map[string]bool) {
	matches := domainRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 {
			domainMap[match[1]] = true
		}
	}
}

// extractIPsFromLine extracts IP addresses from the line.
func extractIPsFromLine(line string, ipMap map[string]bool) {
	matches := ipRegex.FindAllString(line, -1)
	for _, ip := range matches {
		ipMap[ip] = true
	}
}

// extractQueryParamsFromLine extracts query parameters and their values from the line.
func extractQueryParamsFromLine(line string, paramMap map[string]bool) {
	matches := queryParamRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 2 {
			paramMap[fmt.Sprintf("%s=%s", match[1], match[2])] = true
		}
	}
}

// printExtractedData prints the extracted data in sorted order.
// If silent is true, it prints only the values without the header label.
func printExtractedData(label string, dataMap map[string]bool, silent bool) {
	if len(dataMap) == 0 {
		return
	}
	if !silent {
		fmt.Printf("\nExtracted %s:\n", label)
	}
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k)
	}
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
	fmt.Fprintf(w, "        Output data without titles\n\n")
	fmt.Fprintf(w, "Examples:\n")
	fmt.Fprintf(w, "  Extract all patterns:\n")
	fmt.Fprintf(w, "    %s -file input.txt -emails -domains -ips -queryParams\n\n", progName)
	fmt.Fprintf(w, "  Extract only domains and IPs in silent mode:\n")
	fmt.Fprintf(w, "    %s -file input.txt -domains -ips -silent\n\n", progName)
	fmt.Fprintf(w, "  Extract specific UUID version:\n")
	fmt.Fprintf(w, "    %s -file input.txt -uuid 4\n", progName)
}

func main() {
	flag.Usage = func() {
		generateHelpText(os.Stderr, getProgramName())
	}

	// Command-line flags
	filePath := flag.String("file", "", "Path to the input file")
	uuidVersion := flag.Int("uuid", 4, "UUID version to extract (1-5)")
	extractEmails := flag.Bool("emails", false, "Extract email addresses")
	extractDomains := flag.Bool("domains", false, "Extract domain names")
	extractIPs := flag.Bool("ips", false, "Extract IP addresses")
	extractQueryParams := flag.Bool("queryParams", false, "Extract query parameters")
	silent := flag.Bool("silent", false, "Output data without titles")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	// Show help if -help flag is used or no arguments are provided
	if *help || len(os.Args) == 1 {
		flag.Usage()
		return
	}

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: Please provide a file path using -file")
		flag.Usage()
		os.Exit(1)
	}

	// Begin extraction based on the provided flags
	extractData(*filePath, *uuidVersion, *extractEmails, *extractDomains, *extractIPs, *extractQueryParams, *silent)
}
