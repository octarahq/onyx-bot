package utils

import (
	"fmt"
	"strings"
)

func ParseVariables(content string, vars map[string]string) string {
	for key, value := range vars {
		content = strings.ReplaceAll(content, fmt.Sprintf("{%s}", key), value)
	}

	return content
}