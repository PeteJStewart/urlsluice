package redirect

import (
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// RedirectDetector holds configuration for redirect detection
type RedirectDetector struct {
	redirectParams []string
}

// Config represents the YAML configuration structure
type Config struct {
	RedirectParams []string `yaml:"redirect_params"`
}

// Default redirect parameters if no config is provided
var defaultRedirectParams = []string{
	"next",
	"url",
	"redirect",
	"return",
	"goto",
	"dest",
	"view",
}

// NewRedirectDetector creates a new detector with optional configuration
func NewRedirectDetector(configPath string) (*RedirectDetector, error) {
	params := defaultRedirectParams

	if configPath != "" {
		config, err := loadConfig(configPath)
		if err != nil {
			return nil, err
		}
		if len(config.RedirectParams) > 0 {
			params = config.RedirectParams
		}
	}

	return &RedirectDetector{
		redirectParams: params,
	}, nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DetectRedirectParams analyzes a URL for potential open redirect parameters
func (d *RedirectDetector) DetectRedirectParams(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	query := u.Query()
	for param, values := range query {
		// First check known redirect parameters
		isKnownParam := false
		for _, redirectParam := range d.redirectParams {
			if strings.EqualFold(param, redirectParam) {
				isKnownParam = true
				break
			}
		}

		// Check parameter values regardless of parameter name
		for _, value := range values {
			if isURLLike(value) {
				// If it's a known redirect parameter, or if the value is URL-like,
				// consider it a potential redirect
				if isKnownParam || !isNumericOrShort(value) {
					return true
				}
			}
		}
	}

	return false
}

// isURLLike checks if a string looks like a URL
func isURLLike(value string) bool {
	return strings.HasPrefix(value, "http://") ||
		strings.HasPrefix(value, "https://") ||
		strings.HasPrefix(value, "//")
}

// isNumericOrShort returns true if the string is numeric or too short to be a URL
func isNumericOrShort(value string) bool {
	if len(value) < 4 { // too short to be a URL
		return true
	}

	// Check if it's numeric
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// RedirectResult represents the result of scanning a URL for open redirects
type RedirectResult struct {
	URL           string
	IsVulnerable  bool
	MatchedParams []MatchedParameter
}

// MatchedParameter contains details about a matched redirect parameter
type MatchedParameter struct {
	Name    string
	Value   string
	IsKnown bool // Whether it's a known redirect parameter
}

// ScanURLs analyzes multiple URLs for potential open redirects
func (d *RedirectDetector) ScanURLs(urls []string) []RedirectResult {
	results := make([]RedirectResult, 0, len(urls))
	for _, u := range urls {
		result := d.ScanURL(u)
		results = append(results, result)
	}
	return results
}

// ScanURL analyzes a single URL and returns detailed results
func (d *RedirectDetector) ScanURL(urlStr string) RedirectResult {
	result := RedirectResult{
		URL:           urlStr,
		IsVulnerable:  false,
		MatchedParams: make([]MatchedParameter, 0),
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return result
	}

	query := u.Query()
	for param, values := range query {
		// Check if it's a known redirect parameter
		isKnown := false
		for _, redirectParam := range d.redirectParams {
			if strings.EqualFold(param, redirectParam) {
				isKnown = true
				break
			}
		}

		for _, value := range values {
			if isURLLike(value) {
				if isKnown || !isNumericOrShort(value) {
					result.IsVulnerable = true
					result.MatchedParams = append(result.MatchedParams, MatchedParameter{
						Name:    param,
						Value:   value,
						IsKnown: isKnown,
					})
				}
			}
		}
	}

	return result
}
