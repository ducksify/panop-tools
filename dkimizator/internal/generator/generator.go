package generator

import (
	"fmt"
	"strconv"
	"strings"
)

// GenerateSelectors generates all selector combinations from a rule
func GenerateSelectors(rule string, domain string, callback func(string)) error {
	patterns, err := ParsePattern(rule)
	if err != nil {
		return err
	}

	genFunc := func(prefix string) {
		callback(prefix)
	}

	// Build nested generator chain in reverse order
	// This ensures that patterns are applied left-to-right in the final output
	for i := len(patterns) - 1; i >= 0; i-- {
		pattern := patterns[i]
		nextFunc := genFunc
		genFunc = buildGenerator(pattern, domain, nextFunc)
	}

	// Start generation
	genFunc("")
	return nil
}

// buildGenerator creates a generator function for a pattern
func buildGenerator(pattern Pattern, domain string, nextFunc func(string)) func(string) {
	switch pattern.Type {
	case PatternNumeric:
		return func(prefix string) {
			generateNumeric(prefix, pattern.Args, nextFunc)
		}
	case PatternDomain:
		return func(prefix string) {
			generateDomain(prefix, pattern.Args, domain, nextFunc)
		}
	case PatternList:
		return func(prefix string) {
			generateList(prefix, pattern.Args, nextFunc)
		}
	case PatternOptional:
		return func(prefix string) {
			generateOptional(prefix, pattern.Args, nextFunc)
		}
	case PatternString:
		return func(prefix string) {
			nextFunc(prefix + pattern.Raw)
		}
	default:
		return func(prefix string) {
			nextFunc(prefix)
		}
	}
}

// generateNumeric generates numeric range values
func generateNumeric(prefix string, args []string, callback func(string)) {
	if len(args) != 2 {
		return
	}

	start, err1 := strconv.Atoi(args[0])
	end, err2 := strconv.Atoi(args[1])
	if err1 != nil || err2 != nil {
		return
	}

	// Check for zero padding
	zeroPad := strings.HasPrefix(args[0], "0")
	padLen := len(args[0])

	for n := start; n <= end; n++ {
		if zeroPad {
			callback(fmt.Sprintf("%s%0*d", prefix, padLen, n))
		} else {
			callback(fmt.Sprintf("%s%d", prefix, n))
		}
	}
}

// generateDomain generates domain part values
func generateDomain(prefix string, args []string, domain string, callback func(string)) {
	parts := strings.Split(domain, ".")
	numParts := len(parts)

	if len(args) == 0 {
		// Full domain
		callback(prefix + domain)
		return
	}

	if len(args) == 1 {
		// Single part
		idx, err := strconv.Atoi(args[0])
		if err != nil {
			return
		}

		// Convert 1-based to 0-based, handle negative indices
		if idx > 0 {
			idx--
		}
		if idx < 0 {
			idx = numParts + idx
		}

		// Clamp to valid range
		if idx < 0 {
			idx = 0
		}
		if idx >= numParts {
			idx = numParts - 1
		}

		callback(prefix + parts[idx])
		return
	}

	if len(args) == 2 {
		// Range of parts
		start, err1 := strconv.Atoi(args[0])
		end, err2 := strconv.Atoi(args[1])
		if err1 != nil || err2 != nil {
			return
		}

		// Convert 1-based to 0-based, handle negative indices
		if start > 0 {
			start--
		}
		if start < 0 {
			start = numParts + start
		}
		if end > 0 {
			end--
		}
		if end < 0 {
			end = numParts + end
		}

		// Clamp to valid range
		if start < 0 {
			start = 0
		}
		if start >= numParts {
			start = numParts - 1
		}
		if end < 0 {
			end = 0
		}
		if end >= numParts {
			end = numParts - 1
		}

		// Ensure start <= end
		if start > end {
			start, end = end, start
		}

		result := strings.Join(parts[start:end+1], ".")
		callback(prefix + result)
		return
	}
}

// generateList generates values from a list
func generateList(prefix string, args []string, callback func(string)) {
	for _, item := range args {
		callback(prefix + item)
	}
}

// generateOptional generates both with and without the optional string
func generateOptional(prefix string, args []string, callback func(string)) {
	if len(args) != 1 {
		return
	}
	// Generate without
	callback(prefix)
	// Generate with
	callback(prefix + args[0])
}
