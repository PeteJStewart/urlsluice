package redirect

import (
	"os"
	"reflect"
	"testing"
)

func TestDetectRedirectParams(t *testing.T) {
	detector, err := NewRedirectDetector("")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "basic redirect parameter",
			url:      "https://example.com/login?next=https://evil.com",
			expected: true,
		},
		{
			name:     "safe pagination parameter",
			url:      "https://example.com/page?next=2",
			expected: false,
		},
		{
			name:     "no redirect parameter",
			url:      "https://example.com/home",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectRedirectParams(tt.url)
			if result != tt.expected {
				t.Errorf("DetectRedirectParams(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestDetectRedirectParams_AdvancedCases(t *testing.T) {
	detector, err := NewRedirectDetector("")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "double slash bypass",
			url:      "https://example.com/login?goto=//evil.com",
			expected: true,
		},
		{
			name:     "unusual parameter with URL",
			url:      "https://example.com/login?random_param=https://evil.com",
			expected: true,
		},
		{
			name:     "unusual parameter with double slash",
			url:      "https://example.com/login?xyz=//evil.com",
			expected: true,
		},
		{
			name:     "safe unusual parameter",
			url:      "https://example.com/login?random_param=12345",
			expected: false,
		},
		{
			name:     "URL in path segment",
			url:      "https://example.com/login/https://evil.com",
			expected: false, // We only check query parameters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectRedirectParams(tt.url)
			if result != tt.expected {
				t.Errorf("DetectRedirectParams(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

// Test configuration loading
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		wantParams    []string
		wantErr       bool
	}{
		{
			name: "valid config",
			configContent: `redirect_params:
  - next
  - custom_redirect
  - return_url`,
			wantParams: []string{"next", "custom_redirect", "return_url"},
			wantErr:    false,
		},
		{
			name:          "empty config",
			configContent: "",
			wantParams:    defaultRedirectParams,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpfile, err := os.CreateTemp("", "config*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.configContent)); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			detector, err := NewRedirectDetector(tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedirectDetector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(detector.redirectParams, tt.wantParams) {
					t.Errorf("redirectParams = %v, want %v", detector.redirectParams, tt.wantParams)
				}
			}
		})
	}
}

func TestScanURLs(t *testing.T) {
	detector, err := NewRedirectDetector("")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		urls     []string
		expected []RedirectResult
	}{
		{
			name: "multiple URLs with redirects",
			urls: []string{
				"https://example.com/login?next=https://evil.com",
				"https://example.com/page?random=//evil.com",
				"https://example.com/safe?page=2",
			},
			expected: []RedirectResult{
				{
					URL:          "https://example.com/login?next=https://evil.com",
					IsVulnerable: true,
					MatchedParams: []MatchedParameter{
						{
							Name:    "next",
							Value:   "https://evil.com",
							IsKnown: true,
						},
					},
				},
				{
					URL:          "https://example.com/page?random=//evil.com",
					IsVulnerable: true,
					MatchedParams: []MatchedParameter{
						{
							Name:    "random",
							Value:   "//evil.com",
							IsKnown: false,
						},
					},
				},
				{
					URL:           "https://example.com/safe?page=2",
					IsVulnerable:  false,
					MatchedParams: []MatchedParameter{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := detector.ScanURLs(tt.urls)
			if len(results) != len(tt.expected) {
				t.Fatalf("got %d results, want %d", len(results), len(tt.expected))
			}

			for i, result := range results {
				expected := tt.expected[i]
				if result.URL != expected.URL {
					t.Errorf("result[%d].URL = %s, want %s", i, result.URL, expected.URL)
				}
				if result.IsVulnerable != expected.IsVulnerable {
					t.Errorf("result[%d].IsVulnerable = %v, want %v", i, result.IsVulnerable, expected.IsVulnerable)
				}
				if !reflect.DeepEqual(result.MatchedParams, expected.MatchedParams) {
					t.Errorf("result[%d].MatchedParams = %v, want %v", i, result.MatchedParams, expected.MatchedParams)
				}
			}
		})
	}
}
