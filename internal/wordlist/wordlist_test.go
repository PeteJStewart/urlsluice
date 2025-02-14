package wordlist

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"joe_doe-123", []string{"joe", "doe", "123"}},
		{"app.min.js", []string{"app", "min", "js"}},
		{"simple", []string{"simple"}},
		{"", []string{}},
		{"multiple--delimiters__here", []string{"multiple", "delimiters", "here"}},
	}

	for _, tc := range tests {
		result := Tokenize(tc.input)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("Tokenize(%q) = %v; want %v", tc.input, result, tc.expected)
		}
	}
}

func TestGenerateWordlist(t *testing.T) {
	tests := []struct {
		name     string
		urls     []string
		expected []string
	}{
		{
			name: "basic urls",
			urls: []string{
				"https://example.com/path/to/resource",
				"https://example.com/another/path?key=value",
			},
			expected: []string{"another", "key", "path", "resource", "value"},
		},
		{
			name: "handles duplicate words",
			urls: []string{
				"https://example.com/path/to/path",
				"https://example.com/path",
			},
			expected: []string{"path"},
		},
		{
			name: "handles invalid URLs",
			urls: []string{
				"https://example.com/valid/path",
				"://invalid-url",
			},
			expected: []string{"path", "valid"},
		},
		{
			name:     "empty url list",
			urls:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateWordlist(tt.urls)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GenerateWordlist() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractTokensFromURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    []string
		expectError bool
	}{
		{
			name:        "simple path",
			url:         "https://example.com/path/to/resource",
			expected:    []string{"path", "to", "resource"},
			expectError: false,
		},
		{
			name:        "path with query parameters",
			url:         "https://example.com/path?key=value&other=param",
			expected:    []string{"path", "key", "value", "other", "param"},
			expectError: false,
		},
		{
			name:        "empty path",
			url:         "https://example.com",
			expected:    nil,
			expectError: false,
		},
		{
			name:        "invalid URL",
			url:         "://invalid-url",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "URL with encoded parameters",
			url:         "https://example.com/path?key=value%20with%20spaces",
			expected:    []string{"path", "key", "value with spaces"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTokensFromURL(tt.url)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we don't expect an error, check the results
			if !tt.expectError {
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("ExtractTokensFromURL() = %v, want %v", got, tt.expected)
					t.Logf("Length of got: %d, Length of expected: %d", len(got), len(tt.expected))

					// Print each token with its position and exact value
					t.Log("Got tokens:")
					for i, token := range got {
						t.Logf("  [%d] %q", i, token)
					}

					t.Log("Expected tokens:")
					for i, token := range tt.expected {
						t.Logf("  [%d] %q", i, token)
					}

					// Find first difference
					minLen := len(got)
					if len(tt.expected) < minLen {
						minLen = len(tt.expected)
					}
					for i := 0; i < minLen; i++ {
						if got[i] != tt.expected[i] {
							t.Logf("First difference at position %d: got %q, want %q", i, got[i], tt.expected[i])
							break
						}
					}
				}
			}
		})
	}
}
