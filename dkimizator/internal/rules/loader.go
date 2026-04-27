package rules

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Loader handles loading rules from local files or URLs
type Loader struct{}

// NewLoader creates a new rules loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadRules loads rules from a file path or URL
func (l *Loader) LoadRules(source string) ([]string, error) {
	var reader io.Reader

	// Check if it's a URL
	if u, parseErr := url.Parse(source); parseErr == nil && (u.Scheme == "http" || u.Scheme == "https") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch rules from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch rules: HTTP %d", resp.StatusCode)
		}

		reader = resp.Body
	} else {
		// Local file
		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("failed to open rules file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	return l.parseRules(reader)
}

// parseRules parses rules from a reader, handling comments, blank lines, and EoF marker
func (l *Loader) parseRules(reader io.Reader) ([]string, error) {
	var rules []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		// Strip spaces
		line = strings.ReplaceAll(line, " ", "")
		line = strings.ReplaceAll(line, "\t", "")

		// Skip blank lines
		if line == "" {
			continue
		}

		// Skip comments (lines starting with # or ;)
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Stop at EoF marker
		if line == "EoF" {
			break
		}

		rules = append(rules, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading rules: %w", err)
	}

	return rules, nil
}
