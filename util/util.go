package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func QuoteIfNeeded(s string) string {
	if len(s) > 1 && (s[0] == '"' && s[len(s)-1] == '"') {
		return s
	}
	return strconv.Quote(s)
}

func ByteCountIEC(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// checks if the current level is within the wanted slice via a case insensitive comparison.
//
// an empty wanted level slice is the same as specifying all levels.
func IsWantedLevel(wanted []string, current string) bool {
	// specifying no wanted level will default to all levels, return true
	if len(wanted) == 0 {
		return true
	}

	// expand 'err' to 'error'
	if current == "err" {
		current = "error"
	}

	// loop through wanted levels searching via a case insensitive match for either 'ALL' or the current level
	for i := range wanted {
		if strings.EqualFold(wanted[i], "ALL") || strings.EqualFold(wanted[i], current) {
			return true
		}
	}

	return false
}

// New function to check if a log matches the content filter
func MatchesContentFilter(filter string, logContent string) bool {
	if filter == "" {
		return true
	}

	re, err := regexp.Compile(filter)
	if err != nil {
		// If the regex is invalid, treat it as a plain text search
		return strings.Contains(logContent, filter)
	}

	return re.MatchString(logContent)
}
