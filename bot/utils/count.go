package utils

import (
	"fmt"
	"math"
	"strings"
)

func ParseCount(count int) string {
	var result string

	if 0 <= count && count <= 900 {
		result = fmt.Sprintf("%d", count)
	} else if 900 < count && count <= 1000 {
		result = fmt.Sprintf("%.1f", math.Floor((float64(count)/1000)*10)/10)
	} else if 1000 < count && count <= 1000000 {
		result = fmt.Sprintf("%.1fK", math.Floor((float64(count)/1000)*10)/10)
	} else if 1000000 < count {
		result = fmt.Sprintf("%.1fM", math.Floor((float64(count)/1000000)*10)/10)
	} else {
		result = "N/A"
	}

	return result
}

func ParseUnit(b int64, unit string) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	unit = strings.ToUpper(unit)

	if unit == "AUTO" {
		if b == 0 {
			return "0 B"
		}
		
		val := float64(b)
		idx := 0
		for val >= 1024 && idx < len(units)-1 {
			val /= 1024
			idx++
		}
		
		if idx == 0 {
			return fmt.Sprintf("%d B", int64(val))
		}
		return fmt.Sprintf("%.2f %s", val, units[idx])
	}

	idx := 0
	for i, u := range units {
		if u == unit {
			idx = i
			break
		}
	}

	if idx == 0 {
		return fmt.Sprintf("%d B", b)
	}

	val := float64(b) / math.Pow(1024, float64(idx))
	return fmt.Sprintf("%.2f %s", val, units[idx])
}