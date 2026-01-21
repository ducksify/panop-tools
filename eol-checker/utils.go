package main

import (
	"strings"
	"time"
)

func normalizeVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return version
}

func isSupported(eol interface{}, currentDate time.Time) bool {
	if eol == false {
		return true
	}
	eolStr, ok := eol.(string)
	if !ok {
		return false
	}
	if eolStr == "" || eolStr == "unknown" {
		return false
	}
	eolDate, err := time.Parse("2006-01-02", eolStr)
	if err != nil {
		return false
	}
	return eolDate.After(currentDate) || eolDate.Equal(currentDate)
}
