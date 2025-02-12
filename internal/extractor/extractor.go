// Package extractor provides functionality to extract various patterns from text input
package extractor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"urlsluice/internal/patterns"
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

// Extractor defines the interface for pattern extraction
type Extractor interface {
	// Extract processes the input from reader and returns found patterns
	// Context can be used to cancel the operation
	Extract(ctx context.Context, reader io.Reader) (Results, error)
}

// Results contains all patterns found during extraction
type Results struct {
	UUIDs   map[string]bool // Map of unique UUIDs found
	Emails  map[string]bool // Map of unique email addresses
	Domains map[string]bool // Map of unique domain names
	IPs     map[string]bool // Map of unique IP addresses
	Params  map[string]bool // Map of unique query parameters
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
	maxFileSize   = 100 * 1024 * 1024 // 100MB
	chunkSize     = 1 * 1024 * 1024   // 1MB
	maxGoroutines = 4
)

type chunk struct {
	data string
	err  error
}

type extractor struct {
	config Config
}

// New creates a new Extractor with the given configuration
func New(config Config) Extractor {
	return &extractor{
		config: config,
	}
}

func (e *extractor) processChunk(ctx context.Context, data string) Results {
	results := Results{
		UUIDs:   make(map[string]bool),
		Emails:  make(map[string]bool),
		Domains: make(map[string]bool),
		IPs:     make(map[string]bool),
		Params:  make(map[string]bool),
	}

	// Process UUIDs
	if e.config.UUIDVersion > 0 {
		if regex, ok := patterns.UUIDRegexMap[e.config.UUIDVersion]; ok {
			for _, uuid := range regex.FindAllString(data, -1) {
				results.UUIDs[uuid] = true
			}
		}
	}

	// Process other patterns
	if e.config.ExtractEmails {
		for _, email := range patterns.EmailRegex.FindAllString(data, -1) {
			results.Emails[email] = true
		}
	}

	if e.config.ExtractDomains {
		for _, match := range patterns.DomainRegex.FindAllStringSubmatch(data, -1) {
			if len(match) > 1 {
				results.Domains[match[1]] = true
			}
		}
	}

	if e.config.ExtractIPs {
		for _, ip := range patterns.IPRegex.FindAllString(data, -1) {
			if net.ParseIP(ip) != nil {
				results.IPs[ip] = true
			}
		}
	}

	if e.config.ExtractParams {
		for _, match := range patterns.QueryParamRegex.FindAllStringSubmatch(data, -1) {
			if len(match) > 2 {
				results.Params[fmt.Sprintf("%s=%s", match[1], match[2])] = true
			}
		}
	}

	return results
}

func (e *extractor) Extract(ctx context.Context, reader io.Reader) (Results, error) {
	if reader == nil {
		return Results{}, &ExtractorError{Op: "Extract", Err: fmt.Errorf("nil reader")}
	}

	// Check file size
	if f, ok := reader.(*os.File); ok {
		info, err := f.Stat()
		if err != nil {
			return Results{}, &ExtractorError{Op: "Extract", Err: fmt.Errorf("error getting file info: %w", err)}
		}
		if info.Size() > maxFileSize {
			return Results{}, &ExtractorError{Op: "Extract", Err: fmt.Errorf("file too large: maximum size is 100MB")}
		}
	}

	// Create buffered reader
	bufReader := bufio.NewReader(reader)

	// Create channels for chunk processing
	chunks := make(chan chunk, maxGoroutines)
	results := make(chan Results, maxGoroutines)
	errors := make(chan error, 1)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range chunks {
				if c.err != nil {
					select {
					case errors <- c.err:
					default:
					}
					return
				}
				select {
				case results <- e.processChunk(ctx, c.data):
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Read and send chunks
	go func() {
		defer close(chunks)
		buffer := make([]byte, chunkSize)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := bufReader.Read(buffer)
				if n > 0 {
					chunks <- chunk{data: string(buffer[:n])}
				}
				if err == io.EOF {
					return
				}
				if err != nil {
					chunks <- chunk{err: err}
					return
				}
			}
		}
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Combine results
	finalResults := Results{
		UUIDs:   make(map[string]bool),
		Emails:  make(map[string]bool),
		Domains: make(map[string]bool),
		IPs:     make(map[string]bool),
		Params:  make(map[string]bool),
	}

	for {
		select {
		case err := <-errors:
			if err != nil {
				return Results{}, &ExtractorError{Op: "Extract", Err: err}
			}
		case r, ok := <-results:
			if !ok {
				return finalResults, nil
			}
			// Merge results
			for k, v := range r.UUIDs {
				finalResults.UUIDs[k] = v
			}
			for k, v := range r.Emails {
				finalResults.Emails[k] = v
			}
			for k, v := range r.Domains {
				finalResults.Domains[k] = v
			}
			for k, v := range r.IPs {
				finalResults.IPs[k] = v
			}
			for k, v := range r.Params {
				finalResults.Params[k] = v
			}
		case <-ctx.Done():
			return Results{}, &ExtractorError{Op: "Extract", Err: ctx.Err()}
		}
	}
}
