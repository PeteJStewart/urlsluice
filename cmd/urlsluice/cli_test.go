package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)

func TestGetProgramName(t *testing.T) {
	tests := []struct {
		name     string
		osArgs   []string
		expected string
	}{
		{
			name:     "normal binary name",
			osArgs:   []string{"/usr/local/bin/urlsluice"},
			expected: "urlsluice",
		},
		{
			name:     "go run path",
			osArgs:   []string{"/tmp/go-build123/main"},
			expected: "urlsluice",
		},
		{
			name:     "direct main",
			osArgs:   []string{"main"},
			expected: "urlsluice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			os.Args = tt.osArgs
			defer func() { os.Args = oldArgs }()

			got := getProgramName()
			if got != tt.expected {
				t.Errorf("getProgramName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateHelpText(t *testing.T) {
	tests := []struct {
		name         string
		progName     string
		wantContains []string
	}{
		{
			name:     "default help text",
			progName: "urlsluice",
			wantContains: []string{
				"URL Sluice - Extract patterns from text files",
				"Usage: urlsluice [options]",
				"-file string",
				"-uuid int",
				"-emails",
				"-domains",
				"-ips",
				"-queryParams",
				"-silent",
				"Examples:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			generateHelpText(&buf, tt.progName)

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Help text missing expected content: %q", want)
				}
			}
		})
	}
}

func TestRedirectDetection(t *testing.T) {
	// Save original state
	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Restore original state after all tests
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	tests := []struct {
		name          string
		input         string
		args          []string
		wantOutput    []string
		wantErrOutput string
	}{
		{
			name: "basic redirect detection",
			input: `https://example.com/login?next=https://evil.com
https://example.com/page?next=2
https://example.com/goto?redirect=//evil.com`,
			args: []string{"-detect-redirects"},
			wantOutput: []string{
				"Potential Open Redirects:",
				"https://example.com/login?next=https://evil.com",
				"Parameter: next = https://evil.com (Known: true)",
				"https://example.com/goto?redirect=//evil.com",
				"Parameter: redirect = //evil.com (Known: true)",
			},
		},
		{
			name: "redirect detection with silent mode",
			input: `https://example.com/login?next=https://evil.com
https://example.com/goto?redirect=//evil.com`,
			args: []string{"-detect-redirects", "-silent"},
			wantOutput: []string{
				"https://example.com/login?next=https://evil.com",
				"https://example.com/goto?redirect=//evil.com",
			},
		},
		{
			name:  "redirect detection with custom config",
			input: `https://example.com/login?custom=https://evil.com`,
			args:  []string{"-detect-redirects", "-redirect-config", "testdata/redirect.yaml"},
			wantOutput: []string{
				"Potential Open Redirects:",
				"https://example.com/login?custom=https://evil.com",
				"Parameter: custom = https://evil.com (Known: true)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary input file
			tmpfile, err := os.CreateTemp("", "test*.txt")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			// Set up command line arguments
			args := append([]string{"-file", tmpfile.Name()}, tt.args...)
			os.Args = append([]string{"cmd"}, args...)

			// Capture output
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Reset flag.CommandLine to avoid flag redefinition errors
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Run main
			main()

			// Read output
			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check output
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got %q", want, output)
				}
			}

			if tt.wantErrOutput != "" && !strings.Contains(output, tt.wantErrOutput) {
				t.Errorf("error output should contain %q, got %q", tt.wantErrOutput, output)
			}
		})
	}
}
