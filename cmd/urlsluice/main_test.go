package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/PeteJStewart/urlsluice/internal/extractor"
)

// Move osExit to package level
var osExit = os.Exit

// Create a function to restore the original exit function
func restoreExit() {
	osExit = os.Exit
}

func TestPrintResults(t *testing.T) {
	tests := []struct {
		name     string
		results  extractor.Results
		silent   bool
		expected string
	}{
		{
			name: "normal output",
			results: extractor.Results{
				Emails: map[string]bool{
					"test@example.com": true,
					"abc@example.com":  true,
				},
			},
			silent:   false,
			expected: "\nExtracted Emails:\nabc@example.com\ntest@example.com\n",
		},
		{
			name: "silent output",
			results: extractor.Results{
				Emails: map[string]bool{
					"test@example.com": true,
				},
			},
			silent:   true,
			expected: "test@example.com\n",
		},
		{
			name:     "empty results",
			results:  extractor.Results{},
			silent:   false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printResults(tt.results, tt.silent)

			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			os.Stdout = old

			if got := buf.String(); got != tt.expected {
				t.Errorf("printResults() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantConfig  Config
		wantErr     bool
		wantErrText string
	}{
		{
			name: "all flags set",
			args: []string{"-uuid", "4", "-emails", "-domains", "-ips", "-queryParams", "-silent", "-file", "testfile"},
			wantConfig: Config{
				FilePath:       "testfile",
				UUIDVersion:    4,
				ExtractEmails:  true,
				ExtractDomains: true,
				ExtractIPs:     true,
				ExtractParams:  true,
				Silent:         true,
			},
		},
		{
			name:        "missing file",
			args:        []string{"-emails"},
			wantErr:     true,
			wantErrText: "file path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			oldArgs := os.Args
			os.Args = append([]string{"cmd"}, tt.args...)
			defer func() { os.Args = oldArgs }()

			got, err := parseFlags()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErrText) {
				t.Errorf("parseFlags() error = %v, want error containing %q", err, tt.wantErrText)
				return
			}
			if err == nil && !reflect.DeepEqual(got, &tt.wantConfig) {
				t.Errorf("parseFlags() = %v, want %v", got, tt.wantConfig)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		args        []string
		wantErr     bool
		wantErrText string
	}{
		{
			name:    "valid content with emails",
			content: "Contact us at test@example.com or support@example.com",
			args:    []string{"-emails", "-file", "testfile"},
			wantErr: false,
		},
		{
			name:    "valid content with domains",
			content: "Visit https://example.com or http://test.com",
			args:    []string{"-domains", "-file", "testfile"},
			wantErr: false,
		},
		{
			name:    "valid content with IPs",
			content: "Server IPs: 192.168.1.1 and 10.0.0.1",
			args:    []string{"-ips", "-file", "testfile"},
			wantErr: false,
		},
		{
			name:    "valid content with query params",
			content: "URL: https://example.com?param1=value1&param2=value2",
			args:    []string{"-queryParams", "-file", "testfile"},
			wantErr: false,
		},
		{
			name:    "valid content with UUIDs",
			content: "UUID: 550e8400-e29b-41d4-a716-446655440000",
			args:    []string{"-uuid", "4", "-file", "testfile"},
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			args:    []string{"-emails", "-file", "testfile"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpfile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Write content to temp file
			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Set up clean flag set
			oldArgs := os.Args
			oldFlagCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldFlagCommandLine
			}()

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Update args to use the temp file
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "testfile" {
					args[i] = tmpfile.Name()
				}
			}
			os.Args = append([]string{"cmd"}, args...)

			err = run(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrText != "" && !strings.Contains(err.Error(), tt.wantErrText) {
				t.Errorf("run() error = %v, want error containing %q", err, tt.wantErrText)
			}
		})
	}
}

func TestMain_Integration(t *testing.T) {
	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("exit")
	}
	defer restoreExit()

	tests := []struct {
		name       string
		args       []string
		inputFile  string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "valid run with emails",
			args:       []string{"-emails", "-file", "testfile"},
			inputFile:  "test@example.com",
			wantErr:    false,
			wantOutput: "\nExtracted Emails:\ntest@example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			exitCode = 0

			// Reset flags before each test
			oldArgs := os.Args
			oldFlagCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldFlagCommandLine
			}()

			// Set up clean flag set
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = append([]string{"cmd"}, tt.args...)

			// Create a temporary file for the test
			tmpfile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Write the test content
			if _, err := tmpfile.Write([]byte(tt.inputFile)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Update args to use the temp file
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "testfile" {
					args[i] = tmpfile.Name()
				}
			}
			tt.args = args
			os.Args = append([]string{"cmd"}, args...)

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Ensure cleanup
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
				os.Args = os.Args[:len(os.Args)-len(tt.args)]
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			}()

			// Run main with timeout
			done := make(chan bool)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
					w.Close()
					close(done)
				}()
				main()
			}()

			// Wait with timeout
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("Test timed out waiting for completion")
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Verify results
			if (exitCode != 0) != tt.wantErr {
				t.Errorf("Exit code = %d, wantErr %v", exitCode, tt.wantErr)
			}

			if tt.wantOutput != "" && output != tt.wantOutput {
				t.Errorf("main() output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}
