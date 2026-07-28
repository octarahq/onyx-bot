package utils

import "regexp"

var emojiRegex = regexp.MustCompile(`<a?:[a-zA-Z0-9_]+:[0-9]+>`)

func IncludeEmojis(content string) (bool, int) {
	matches := emojiRegex.FindAllString(content, -1)
	return len(matches) > 0, len(matches)
}
