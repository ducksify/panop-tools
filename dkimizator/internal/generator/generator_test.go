package generator

import (
	"sort"
	"testing"
)

func TestGenerateSelectors(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		domain   string
		expected []string
	}{
		{
			name:     "simple string",
			rule:     "dkim",
			domain:   "example.com",
			expected: []string{"dkim"},
		},
		{
			name:     "string with numeric range",
			rule:     "dkim{N:1-9}",
			domain:   "example.com",
			expected: []string{"dkim1", "dkim2", "dkim3", "dkim4", "dkim5", "dkim6", "dkim7", "dkim8", "dkim9"},
		},
		{
			name:     "numeric range with zero padding",
			rule:     "k{N:01-05}",
			domain:   "example.com",
			expected: []string{"k01", "k02", "k03", "k04", "k05"},
		},
		{
			name:     "list pattern",
			rule:     "{L:default,google,mail}",
			domain:   "example.com",
			expected: []string{"default", "google", "mail"},
		},
		{
			name:     "string with list",
			rule:     "s{L:384,512,768}",
			domain:   "example.com",
			expected: []string{"s384", "s512", "s768"},
		},
		{
			name:     "optional pattern",
			rule:     "dkim{O:-test}",
			domain:   "example.com",
			expected: []string{"dkim", "dkim-test"},
		},
		{
			name:     "domain pattern",
			rule:     "{D}",
			domain:   "example.com",
			expected: []string{"example.com"},
		},
		{
			name:     "domain first part",
			rule:     "{D:1}",
			domain:   "example.com",
			expected: []string{"example"},
		},
		{
			name:   "complex pattern",
			rule:   "mail{N:2005-2018}{O:-}{N:01-12}",
			domain: "example.com",
			// Optional pattern generates both with and without, so we expect 2x results
			// This test just verifies the pattern works, not the exact output
			expected: []string{}, // Will check count instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []string
			err := GenerateSelectors(tt.rule, tt.domain, func(selector string) {
				results = append(results, selector)
			})

			if err != nil {
				t.Fatalf("GenerateSelectors() error = %v", err)
			}

			// Sort both slices for comparison
			sort.Strings(results)
			sort.Strings(tt.expected)

			// Special handling for complex pattern test
			if tt.name == "complex pattern" {
				// Should generate 336 results (14 years * 12 months * 2 for optional)
				if len(results) != 336 {
					t.Errorf("GenerateSelectors() count mismatch: got %d, want 336", len(results))
				}
				// Verify some expected patterns exist
				expectedPatterns := []string{"mail200501", "mail2005-01", "mail201812", "mail2018-12"}
				for _, pattern := range expectedPatterns {
					found := false
					for _, r := range results {
						if r == pattern {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected pattern %q not found in results", pattern)
					}
				}
				return
			}

			if len(results) != len(tt.expected) {
				t.Errorf("GenerateSelectors() count mismatch: got %d, want %d", len(results), len(tt.expected))
				t.Errorf("Got: %v", results)
				t.Errorf("Want: %v", tt.expected)
				return
			}

			for i := range results {
				if results[i] != tt.expected[i] {
					t.Errorf("GenerateSelectors() result[%d] = %v, want %v", i, results[i], tt.expected[i])
				}
			}
		})
	}
}
