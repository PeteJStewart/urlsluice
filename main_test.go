package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestData represents test input and expected output
type TestData struct {
	input    string
	expected map[string]bool
}

func createTestFile(t *testing.T, content string) (string, func()) {
	tmpfile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name(), func() {
		os.Remove(tmpfile.Name())
	}
}

func TestExtractUUIDs(t *testing.T) {
	tests := []TestData{
		{
			input: "550e8400-e29b-41d4-a716-446655440000",
			expected: map[string]bool{
				"550e8400-e29b-41d4-a716-446655440000": true,
			},
		},
		{
			input: "invalid-uuid 550e8400-e29b-41d4-a716-446655440000 another-invalid",
			expected: map[string]bool{
				"550e8400-e29b-41d4-a716-446655440000": true,
			},
		},
		{
			input:    "no uuid here",
			expected: map[string]bool{},
		},
	}

	for _, test := range tests {
		result := make(map[string]bool)
		extractUUIDs(test.input, uuidRegexMap[4], result)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}

func TestExtractEmails(t *testing.T) {
	tests := []TestData{
		{
			input: "user@example.com",
			expected: map[string]bool{
				"user@example.com": true,
			},
		},
		{
			input: "text user@example.com more.user@sub.example.com text",
			expected: map[string]bool{
				"user@example.com":          true,
				"more.user@sub.example.com": true,
			},
		},
		{
			input:    "no email here",
			expected: map[string]bool{},
		},
	}

	for _, test := range tests {
		result := make(map[string]bool)
		extractEmailsFromLine(test.input, result)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}

func TestExtractDomains(t *testing.T) {
	tests := []TestData{
		{
			input: "https://example.com",
			expected: map[string]bool{
				"example.com": true,
			},
		},
		{
			input: "text https://example.com http://sub.domain.com text",
			expected: map[string]bool{
				"example.com":    true,
				"sub.domain.com": true,
			},
		},
		{
			input:    "no domain here",
			expected: map[string]bool{},
		},
	}

	for _, test := range tests {
		result := make(map[string]bool)
		extractDomainsFromLine(test.input, result)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}

func TestIntegration(t *testing.T) {
	content := `https://example.com/users?id=123&token=abc
user@example.com
192.168.1.1
550e8400-e29b-41d4-a716-446655440000`

	filepath, cleanup := createTestFile(t, content)
	defer cleanup()

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Run the extraction
	extractData(filepath, 4, true, true, true, true, false)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Test that the file was processed without errors
	// In a real test, you might want to capture and verify the actual output
	// but for this example, we're just ensuring it runs without panicking
}

func TestHelpOutput(t *testing.T) {
	t.Run("contains essential help sections", func(t *testing.T) {
		var buf bytes.Buffer
		generateHelpText(&buf, "urlsluice")
		output := buf.String()

		// Test for presence of major sections
		essentialSections := []struct {
			name     string
			expected bool
		}{
			{"title", strings.Contains(output, "URL Sluice")},
			{"usage", strings.Contains(output, "Usage:")},
			{"options", strings.Contains(output, "Options:")},
			{"examples", strings.Contains(output, "Examples:")},
		}

		for _, section := range essentialSections {
			if !section.expected {
				t.Errorf("Help text missing essential section: %s", section.name)
			}
		}
	})

	t.Run("documents all flags", func(t *testing.T) {
		var buf bytes.Buffer
		generateHelpText(&buf, "urlsluice")
		output := buf.String()

		requiredFlags := []string{
			"-file",
			"-uuid",
			"-emails",
			"-domains",
			"-ips",
			"-queryParams",
			"-silent",
		}

		for _, flag := range requiredFlags {
			if !strings.Contains(output, flag) {
				t.Errorf("Help text missing documentation for flag: %s", flag)
			}
		}
	})

	t.Run("program name is used consistently", func(t *testing.T) {
		var buf bytes.Buffer
		testProgramName := "test-program"
		generateHelpText(&buf, testProgramName)
		output := buf.String()

		if !strings.Contains(output, fmt.Sprintf("Usage: %s", testProgramName)) {
			t.Error("Help text doesn't use provided program name in usage section")
		}

		// Check that the program name appears in examples
		expectedExampleCount := 3 // We have 3 examples in our help text
		actualCount := strings.Count(output, testProgramName)
		if actualCount < expectedExampleCount {
			t.Errorf("Expected program name to appear at least %d times in examples, but found %d",
				expectedExampleCount, actualCount)
		}
	})
}

// Test the integration with flag package
func TestHelpFlag(t *testing.T) {
	// Save original flags and args
	oldFlagCommandLine := flag.CommandLine
	oldArgs := os.Args
	defer func() {
		flag.CommandLine = oldFlagCommandLine
		os.Args = oldArgs
	}()

	// Set up test flags
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)

	// Set up the usage function
	flag.Usage = func() {
		generateHelpText(&buf, "urlsluice")
	}

	// Test with -help flag
	os.Args = []string{"urlsluice", "-help"}
	if err := flag.CommandLine.Parse(os.Args[1:]); err != flag.ErrHelp {
		t.Error("Expected -help flag to trigger help output")
	}

	if buf.Len() == 0 {
		t.Error("No help text was generated when -help flag was used")
	}
}

func TestGetProgramName(t *testing.T) {
	tests := []struct {
		arg      string
		expected string
	}{
		{"/tmp/go-build123/whatever/main", "urlsluice"},
		{"urlsluice", "urlsluice"},
		{"/usr/local/bin/urlsluice", "urlsluice"},
		{"main", "urlsluice"},
	}

	for _, test := range tests {
		oldArgs := os.Args
		os.Args = []string{test.arg}

		result := getProgramName()
		if result != test.expected {
			t.Errorf("getProgramName() with arg '%s': expected '%s', got '%s'",
				test.arg, test.expected, result)
		}

		os.Args = oldArgs
	}
}
