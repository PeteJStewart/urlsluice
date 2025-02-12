package patterns

import "regexp"

var (
	UUIDRegexMap = map[int]*regexp.Regexp{
		1: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-1[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		2: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-2[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		3: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-3[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		4: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
		5: regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-5[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`),
	}

	EmailRegex      = regexp.MustCompile(`[\w._%+-]+@[\w.-]+\.[a-zA-Z]{2,}`)
	DomainRegex     = regexp.MustCompile(`https?://([a-zA-Z0-9.-]+)/?`)
	IPRegex         = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	QueryParamRegex = regexp.MustCompile(`[?&]([^&=]+)=([^&=]*)`)
)
