package wordlist

import (
	"net/url"
	"sort"
	"strings"
)

func GenerateWordlist(urls []string) []string {
	wordSet := make(map[string]struct{})
	for _, urlStr := range urls {
		tokens, err := ExtractTokensFromURL(urlStr)
		if err != nil {
			continue
		}
		for _, token := range tokens {
			if IsUsefulToken(token) {
				wordSet[strings.ToLower(token)] = struct{}{}
			}
		}
	}
	words := make([]string, 0, len(wordSet))
	for w := range wordSet {
		words = append(words, w)
	}
	sort.Strings(words)
	return words
}

func ExtractTokensFromURL(urlStr string) ([]string, error) {
	var tokens []string
	u, err := url.Parse(urlStr)
	if err != nil {
		return tokens, err
	}
	segments := strings.Split(u.Path, "/")
	for _, segment := range segments {
		if segment != "" {
			tokens = append(tokens, Tokenize(segment)...)
		}
	}
	queryParams := u.Query()
	for key, values := range queryParams {
		tokens = append(tokens, Tokenize(key)...)
		for _, value := range values {
			tokens = append(tokens, Tokenize(value)...)
		}
	}
	return tokens, nil
}
