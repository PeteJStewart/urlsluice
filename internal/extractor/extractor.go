// Package extractor provides functionality for extracting and validating various patterns from text input.
// It supports concurrent processing of large files while maintaining memory efficiency through chunked processing.
// Supported patterns include UUIDs, email addresses, domain names, IP addresses, and URL query parameters.
package extractor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/PeteJStewart/urlsluice/internal/patterns"
)

// ExtractorError represents an error that occurred during extraction
type ExtractorError struct {
	Op  string
	Err error
}

func (e *ExtractorError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ExtractorError) Unwrap() error {
	return e.Err
}

// Extractor defines the interface for pattern extraction operations.
// Implementations must support concurrent processing and respect context cancellation.
type Extractor interface {
	// Extract processes the input from reader and returns found patterns.
	// It supports concurrent processing and respects context cancellation.
	// Returns ExtractorError if processing fails or context is cancelled.
	Extract(ctx context.Context, reader io.Reader) (Results, error)
}

// Results contains all patterns found during extraction.
// Each field is a map using the pattern as key and a boolean as value to ensure uniqueness.
type Results struct {
	// UUIDs stores unique Universal Unique Identifiers
	UUIDs map[string]bool
	// Emails stores unique email addresses
	Emails map[string]bool
	// Domains stores unique domain names extracted from URLs
	Domains map[string]bool
	// IPs stores unique IPv4 addresses
	IPs map[string]bool
	// Params stores unique URL query parameters in "key=value" format
	Params map[string]bool
}

// Config defines the configuration for pattern extraction
type Config struct {
	UUIDVersion    int  // Version of UUIDs to extract (1-5)
	ExtractEmails  bool // Whether to extract email addresses
	ExtractDomains bool // Whether to extract domain names
	ExtractIPs     bool // Whether to extract IP addresses
	ExtractParams  bool // Whether to extract query parameters
}

const (
	// maxFileSize defines the maximum allowed file size (100MB) to prevent memory exhaustion
	maxFileSize = 100 * 1024 * 1024
	// chunkSize defines the size of each processing chunk (1MB) for optimal performance
	chunkSize = 1 * 1024 * 1024
	// maxGoroutines defines the maximum number of concurrent workers
	maxGoroutines = 4
)

type chunk struct {
	data string
	err  error
}

type extractor struct {
	config Config
}

// New creates a new Extractor with the given configuration.
// It validates the configuration and returns an error if:
// - UUID version is not between 0 and 5 (0 disables UUID extraction)
// Returns an initialized Extractor and nil error if configuration is valid.
func New(config Config) (Extractor, error) {
	if config.UUIDVersion < 0 || config.UUIDVersion > 5 {
		return nil, &ExtractorError{Op: "New", Err: fmt.Errorf("invalid UUID version: must be between 0 and 5")}
	}
	return &extractor{
		config: config,
	}, nil
}

func (e *extractor) newResults() Results {
	return Results{}
}

func (e *extractor) processChunk(ctx context.Context, data string) Results {
	select {
	case <-ctx.Done():
		return Results{}
	default:
	}

	results := Results{}
	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		if e.config.UUIDVersion > 0 {
			if regex, ok := patterns.UUIDRegexMap[e.config.UUIDVersion]; ok {
				matches := regex.FindAllString(line, -1)
				if len(matches) > 0 {
					if results.UUIDs == nil {
						results.UUIDs = make(map[string]bool)
					}
					for _, uuid := range matches {
						results.UUIDs[uuid] = true
					}
				}
			}
		}

		if e.config.ExtractEmails {
			matches := patterns.EmailRegex.FindAllString(line, -1)
			if len(matches) > 0 {
				if results.Emails == nil {
					results.Emails = make(map[string]bool)
				}
				for _, email := range matches {
					results.Emails[email] = true
				}
			}
		}

		if e.config.ExtractDomains {
			matches := patterns.DomainRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 && !strings.HasPrefix(match[1], ".") && !strings.HasSuffix(match[1], ".") {
					if results.Domains == nil {
						results.Domains = make(map[string]bool)
					}
					results.Domains[match[1]] = true
				}
			}
		}

		if e.config.ExtractIPs {
			for _, ip := range patterns.IPRegex.FindAllString(line, -1) {
				if net.ParseIP(ip) != nil {
					if results.IPs == nil {
						results.IPs = make(map[string]bool)
					}
					results.IPs[ip] = true
				}
			}
		}

		if e.config.ExtractParams {
			matches := patterns.QueryParamRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 2 {
					if results.Params == nil {
						results.Params = make(map[string]bool)
					}
					results.Params[match[1]+"="+match[2]] = true
				}
			}
		}
	}

	return results
}

func (e *extractor) Extract(ctx context.Context, reader io.Reader) (Results, error) {
	// First, check context before doing anything
	if ctx.Err() != nil {
		return e.newResults(), &ExtractorError{Op: "Extract", Err: ctx.Err()}
	}

	if reader == nil {
		return e.newResults(), &ExtractorError{Op: "Extract", Err: fmt.Errorf("nil reader")}
	}

	// Check file size
	if f, ok := reader.(*os.File); ok {
		info, err := f.Stat()
		if err != nil {
			return e.newResults(), &ExtractorError{Op: "Extract", Err: fmt.Errorf("error getting file info: %w", err)}
		}
		if info.Size() > maxFileSize {
			return e.newResults(), &ExtractorError{Op: "Extract", Err: fmt.Errorf("file too large: maximum size is 100MB")}
		}
	}

	chunks := make(chan chunk, maxGoroutines)
	results := make(chan Results, maxGoroutines)
	errors := make(chan error, 1)

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < maxGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range chunks {
				select {
				case <-ctx.Done():
					select {
					case errors <- ctx.Err():
					default:
					}
					return
				default:
					if c.err != nil {
						select {
						case errors <- c.err:
						default:
						}
						return
					}
					results <- e.processChunk(ctx, c.data)
				}
			}
		}()
	}

	// Read chunks
	go func() {
		defer close(chunks)
		buffer := make([]byte, chunkSize)
		for {
			select {
			case <-ctx.Done():
				chunks <- chunk{err: ctx.Err()} // Send context error through chunks
				return
			default:
				n, err := reader.Read(buffer)
				if err != nil && err != io.EOF {
					chunks <- chunk{err: err}
					return
				}
				if n > 0 {
					chunks <- chunk{data: string(buffer[:n])}
				}
				if err == io.EOF {
					return
				}
			}
		}
	}()

	// Close results after workers finish
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	finalResults := e.newResults()

	// Process results and errors
	for {
		select {
		case err := <-errors:
			if err != nil {
				return e.newResults(), &ExtractorError{Op: "Extract", Err: err}
			}
		case r, ok := <-results:
			if !ok {
				return finalResults, nil
			}
			// Merge results
			if r.UUIDs != nil && len(r.UUIDs) > 0 {
				if finalResults.UUIDs == nil {
					finalResults.UUIDs = make(map[string]bool)
				}
				for k, v := range r.UUIDs {
					finalResults.UUIDs[k] = v
				}
			}
			if r.Emails != nil && len(r.Emails) > 0 {
				if finalResults.Emails == nil {
					finalResults.Emails = make(map[string]bool)
				}
				for k, v := range r.Emails {
					finalResults.Emails[k] = v
				}
			}
			if r.Domains != nil && len(r.Domains) > 0 {
				if finalResults.Domains == nil {
					finalResults.Domains = make(map[string]bool)
				}
				for k, v := range r.Domains {
					finalResults.Domains[k] = v
				}
			}
			if r.IPs != nil && len(r.IPs) > 0 {
				if finalResults.IPs == nil {
					finalResults.IPs = make(map[string]bool)
				}
				for k, v := range r.IPs {
					finalResults.IPs[k] = v
				}
			}
			if r.Params != nil && len(r.Params) > 0 {
				if finalResults.Params == nil {
					finalResults.Params = make(map[string]bool)
				}
				for k, v := range r.Params {
					finalResults.Params[k] = v
				}
			}
		case <-ctx.Done():
			return e.newResults(), &ExtractorError{Op: "Extract", Err: ctx.Err()}
		}
	}
}
