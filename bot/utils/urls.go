package utils

import (
	"net/url"
	"regexp"
	"strings"
)

var UrlRegex = regexp.MustCompile(`https?://[^\s<>"'\(\)]+`)
var MDURLRegex = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)<>"']+)\)`)

func ExtractURLs(content string) []*url.URL {
	rawURLs := UrlRegex.FindAllString(content, -1)
	if len(rawURLs) == 0 {
		return nil
	}

	var parsedURLs []*url.URL

	for _, raw := range rawURLs {
		parsed, err := url.Parse(raw)
		if err == nil && parsed.Scheme != "" && parsed.Host != "" {
			parsedURLs = append(parsedURLs, parsed)
		}
	}

	return parsedURLs
}

type MDURL struct {
	Name string
	URL  url.URL
}

func ExtractMDURLs(content string) []MDURL {
	matches := MDURLRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	var parsedURLs []MDURL

	for _, match := range matches {
		parsed, err := url.Parse(match[2])
		if err == nil && parsed.Scheme != "" && parsed.Host != "" {
			parsedURLs = append(parsedURLs, MDURL{Name: match[1], URL: *parsed})
		}
	}

	return parsedURLs
}

func ExtractDomain(text string) string {
	words := strings.Fields(text)

	for _, word := range words {
		word = strings.ToLower(word)
		word = strings.TrimPrefix(word, "http://")
		word = strings.TrimPrefix(word, "https://")

		host, _, _ := strings.Cut(word, "/")

		host = strings.TrimPrefix(host, "www.")

		if strings.Contains(host, ".") && !strings.HasPrefix(host, ".") && !strings.HasSuffix(host, ".") {
			return host
		}
	}

	return ""
}
