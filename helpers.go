// SPDX-License-Identifier: MIT OR Unlicense

package main

import "time"

// Returns the current time as a millisecond timestamp
func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Returns the current time as a nanosecond timestamp as some things
// are far too fast to measure using nanoseconds
func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
