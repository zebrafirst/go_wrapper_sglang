package utils

import "time"

// Get millis timestamp.
func CurrentTimeMillis() int64 {
	return int64(time.Now().UnixNano() / int64(time.Millisecond))
}
