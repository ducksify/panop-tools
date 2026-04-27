package generator

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternType represents the type of pattern
type PatternType int

const (
	PatternNumeric PatternType = iota
	PatternDomain
	PatternList
	PatternOptional
	PatternString
)

// Pattern represents a parsed pattern element
type Pattern struct {
	Type     PatternType
	Raw      string
	Args     []string
	Original string
}

// ParsePattern parses a pattern string into Pattern elements
// Pattern syntax:
//   - {N:1-20} - numeric range (1 to 20)
//   - {N:01-20} - numeric range with zero padding
//   - {D} - full domain
//   - {D:1} - first domain part
//   - {D:1-3} - domain parts 1 through 3
//   - {D:-1} - last domain part
//   - {D:-3--1} - last three parts
//   - {L:a,b,c} - list of strings
//   - {O:foo} - optional string (generates both with and without)
func ParsePattern(rule string) ([]Pattern, error) {
	var patterns []Pattern
	patternRegex := regexp.MustCompile(`\{([A-Z])([^}]*)\}`)

	lastIndex := 0
	matches := patternRegex.FindAllStringSubmatchIndex(rule, -1)

	for _, match := range matches {
		// Add string before pattern
		if match[0] > lastIndex {
			text := rule[lastIndex:match[0]]
			if text != "" {
				patterns = append(patterns, Pattern{
					Type:     PatternString,
					Raw:      text,
					Original: text,
				})
			}
		}

		// Parse the pattern
		patternType := rule[match[2]:match[3]]
		patternArgs := rule[match[4]:match[5]]
		fullPattern := rule[match[0]:match[1]]

		var pType PatternType
		var args []string

		switch patternType {
		case "N":
			pType = PatternNumeric
			// Strip leading colon if present
			if strings.HasPrefix(patternArgs, ":") {
				patternArgs = patternArgs[1:]
			}
			args = parseNumericArgs(patternArgs)
		case "D":
			pType = PatternDomain
			// Strip leading colon if present
			if strings.HasPrefix(patternArgs, ":") {
				patternArgs = patternArgs[1:]
			}
			args = parseDomainArgs(patternArgs)
		case "L":
			pType = PatternList
			// Strip leading colon if present
			if strings.HasPrefix(patternArgs, ":") {
				patternArgs = patternArgs[1:]
			}
			args = parseListArgs(patternArgs)
		case "O":
			pType = PatternOptional
			// Strip leading colon if present
			if strings.HasPrefix(patternArgs, ":") {
				patternArgs = patternArgs[1:]
			}
			args = []string{patternArgs}
		default:
			return nil, fmt.Errorf("unknown pattern type: %s", patternType)
		}

		patterns = append(patterns, Pattern{
			Type:     pType,
			Raw:      fullPattern,
			Args:     args,
			Original: fullPattern,
		})

		lastIndex = match[1]
	}

	// Add remaining string after last pattern
	if lastIndex < len(rule) {
		text := rule[lastIndex:]
		if text != "" {
			patterns = append(patterns, Pattern{
				Type:     PatternString,
				Raw:      text,
				Original: text,
			})
		}
	}

	// If no patterns found, treat entire rule as string
	if len(patterns) == 0 {
		patterns = append(patterns, Pattern{
			Type:     PatternString,
			Raw:      rule,
			Original: rule,
		})
	}

	return patterns, nil
}

func parseNumericArgs(args string) []string {
	parts := strings.Split(args, "-")
	if len(parts) != 2 {
		return []string{}
	}
	return []string{parts[0], parts[1]}
}

func parseDomainArgs(args string) []string {
	if args == "" {
		return []string{}
	}
	parts := strings.Split(args, "-")
	return parts
}

func parseListArgs(args string) []string {
	if args == "" {
		return []string{}
	}
	// Handle empty first element (leading comma)
	parts := strings.Split(args, ",")
	if len(parts) > 0 && parts[0] == "" {
		// Leading comma - first element is empty string
		return parts
	}
	return parts
}
