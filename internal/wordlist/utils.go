package wordlist

import (
	"net"
	"regexp"
	"strings"
)

var (
	uuidRegex  = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	emailRegex = regexp.MustCompile(`^[\w\.\-]+@[\w\.\-]+\.\w+$`)
)

func Tokenize(input string) []string {
	return strings.FieldsFunc(input, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || r == '/'
	})
}

func IsUsefulToken(token string) bool {
	token = strings.TrimSpace(token)
	if len(token) < 3 || len(token) > 50 {
		return false
	}
	if uuidRegex.MatchString(token) || emailRegex.MatchString(token) {
		return false
	}
	if net.ParseIP(token) != nil {
		return false
	}
	if isNumeric(token) {
		return false
	}
	return true
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
