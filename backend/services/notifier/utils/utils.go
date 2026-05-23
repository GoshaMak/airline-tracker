package utils

import "time"

func InTimeSpan(start, end, check time.Time) bool {
	return (check.After(start) || check.Equal(start)) &&
		(check.Before(end) || check.Equal(end))
}
