package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRegex = regexp.MustCompile(`^(\d+)\s*([a-zA-Z]+)$`)

func ParseDurationToTime(duration string) (time.Time, error) {
	matches := durationRegex.FindStringSubmatch(strings.TrimSpace(duration))
	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("Invalid format duration : %s", duration)
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, err
	}

	unit := strings.ToLower(matches[2])
	now := time.Now()

	switch unit {
	case "s", "sec", "second", "seconds":
		return now.Add(time.Duration(amount) * time.Second), nil
	case "m", "min", "minute", "minutes":
		return now.Add(time.Duration(amount) * time.Minute), nil
	case "h", "hr", "hour", "hours":
		return now.Add(time.Duration(amount) * time.Hour), nil
	case "d", "day", "days":
		return now.AddDate(0, 0, amount), nil
	case "w", "wk", "week", "weeks":
		return now.AddDate(0, 0, amount*7), nil
	case "mo", "month", "months":
		return now.AddDate(0, amount, 0), nil
	case "y", "yr", "year", "years":
		return now.AddDate(amount, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("Unknown unit : %s", unit)
	}
}
