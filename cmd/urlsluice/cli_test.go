package main

import (
	"bytes"
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
