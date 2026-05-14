package util

import "strings"

// ErrorDetect reports whether an API response string appears to contain an error.
func ErrorDetect(s string) bool {
	return strings.Contains(strings.ToLower(s), "error")
}
