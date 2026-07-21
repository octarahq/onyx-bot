package utils

import (
	"fmt"
	"math"
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
