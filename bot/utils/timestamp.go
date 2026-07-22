package utils

import "fmt"

type TimestampFormat string

const (
	TimestampShortTime     TimestampFormat = "t"
	TimestampLongTime      TimestampFormat = "T"
	TimestampShortDate     TimestampFormat = "d"
	TimestampLongDate      TimestampFormat = "D"
	TimestampShortDateTime TimestampFormat = "f"
	TimestampLongDateTime  TimestampFormat = "F"
	TimestampRelativeTime  TimestampFormat = "R"
)

func GenerateTimestamp(timestamp int, format TimestampFormat) string {
	return fmt.Sprintf("<t:%d:%s>", timestamp, string(format))
}
