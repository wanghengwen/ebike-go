package utils

import "strconv"

// Itoa formats int as decimal string without allocation for common paths.
func Itoa(n int) string {
	return strconv.Itoa(n)
}
