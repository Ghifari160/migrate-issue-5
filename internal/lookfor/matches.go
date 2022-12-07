package lookfor

import (
	"runtime"
	"strings"
)

// NameMatches checks if two file name matches. It assumes case-insensitivity on Windows.
func NameMatches(expected, found string) bool {
	if runtime.GOOS == "windows" {
		expected = strings.ToLower(expected)
		found = strings.ToLower(found)
	}

	return expected == found
}
