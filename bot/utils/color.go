package utils

import (
	"strconv"
	"strings"
)

func ParseStrColor(scolor string) int {
	scolor = strings.ReplaceAll(scolor, "#", "")
	color, err := strconv.ParseInt(scolor, 16, 32)
	if err != nil {
		return 0
	}
	return int(color)
}
