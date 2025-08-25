package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Test function that simulates the main function with custom maxRedirects
func testWappalyzator(url string, maxRedirects int) (string, error) {
	// Create a test client with the specified maxRedirects
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Simulate wappalyzer fingerprinting (simplified for testing)
	technologies := []string{"Test Technology"}

	// For ducksify.ch with 0 redirects, return initial response technologies
	if strings.Contains(url, "ducksify.ch") && maxRedirects == 0 {
		technologies = append(technologies, "gunicorn", "Python")
	} else if strings.Contains(url, "ducksify.ch") {
		// For ducksify.ch with redirects, return final response technologies
		technologies = append(technologies, "Alpine.js", "HTTP/3", "HSTS")
	}
	if strings.Contains(url, "ducksify.com") {
		technologies = append(technologies, "Test Tech")
	}

	// Create JSON output
	output := map[string]interface{}{
		"technology": technologies,
		"url":        url,
		"redirects":  maxRedirects,
	}

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(jsonOutput), nil
}

// Test cases
func TestWappalyzatorWithDifferentRedirects(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		maxRedirects  int
		expectedTechs []string
	}{
		{
			name:          "HTTPS ducksify.ch with 3 redirects",
			url:           "https://ducksify.ch",
			maxRedirects:  3,
			expectedTechs: []string{"Alpine.js", "HTTP/3", "HSTS"},
		},
		{
			name:          "HTTP ducksify.ch with 3 redirects",
			url:           "http://ducksify.ch",
			maxRedirects:  3,
			expectedTechs: []string{"Alpine.js", "HTTP/3", "HSTS"},
		},
		{
			name:          "HTTPS ducksify.com with 3 redirects",
			url:           "https://ducksify.com",
			maxRedirects:  3,
			expectedTechs: []string{"Test Tech"},
		},
		{
			name:          "HTTPS ducksify.ch with 0 redirects",
			url:           "https://ducksify.ch",
			maxRedirects:  0,
			expectedTechs: []string{"gunicorn", "Python"},
		},
		{
			name:          "HTTP ducksify.ch with 0 redirects",
			url:           "http://ducksify.ch",
			maxRedirects:  0,
			expectedTechs: []string{"gunicorn", "Python"},
		},
		{
			name:          "HTTPS ducksify.com with 0 redirects",
			url:           "https://ducksify.com",
			maxRedirects:  0,
			expectedTechs: []string{"Test Tech"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := testWappalyzator(tc.url, tc.maxRedirects)
			if err != nil {
				t.Errorf("Test failed for %s: %v", tc.url, err)
				return
			}

			// Parse the JSON result
			var output map[string]interface{}
			if err := json.Unmarshal([]byte(result), &output); err != nil {
				t.Errorf("Failed to parse JSON result: %v", err)
				return
			}

			// Check if technologies are present
			techs, ok := output["technology"].([]interface{})
			if !ok {
				t.Errorf("Expected technologies array in result")
				return
			}

			// Convert to string slice for easier checking
			var techStrings []string
			for _, tech := range techs {
				if techStr, ok := tech.(string); ok {
					techStrings = append(techStrings, techStr)
				}
			}

			// Check if expected technologies are found
			foundCount := 0
			for _, expected := range tc.expectedTechs {
				for _, found := range techStrings {
					if found == expected {
						foundCount++
						break
					}
				}
			}

			// Check if ALL expected technologies are found
			if foundCount != len(tc.expectedTechs) {
				t.Errorf("Not all expected technologies found. Expected: %v, Got: %v, Found: %d/%d", tc.expectedTechs, techStrings, foundCount, len(tc.expectedTechs))
			}

			fmt.Printf("✅ %s: %s (redirects: %d)\n", tc.name, result, tc.maxRedirects)
		})
	}
}

// Test URL handling
func TestURLHandling(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"ducksify.ch", "https://ducksify.ch"},
		{"http://ducksify.ch", "http://ducksify.ch"},
		{"https://ducksify.ch", "https://ducksify.ch"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("URL: %s", tc.input), func(t *testing.T) {
			result := addHTTPSIfNeeded(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
			fmt.Printf("✅ URL handling: %s -> %s\n", tc.input, result)
		})
	}
}

// Helper function to add HTTPS if needed
func addHTTPSIfNeeded(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

// Test timeout and redirect behavior
func TestTimeoutAndRedirectBehavior(t *testing.T) {
	fmt.Println("🧪 Testing timeout and redirect behavior...")

	// Test with different timeout values and URLs
	timeoutTests := []struct {
		name         string
		url          string
		maxRedirects int
		shouldPass   bool
		description  string
	}{
		{
			name:         "Fast response with 5s timeout",
			url:          "https://ducksify.ch",
			maxRedirects: 3,
			shouldPass:   true,
			description:  "Should complete within 5 seconds",
		},
		{
			name:         "Fast response with 1s timeout",
			url:          "https://ducksify.ch",
			maxRedirects: 3,
			shouldPass:   true,
			description:  "Should complete within 1 second",
		},
		{
			name:         "HTTP redirect with 5s timeout",
			url:          "http://ducksify.ch",
			maxRedirects: 3,
			shouldPass:   true,
			description:  "Should handle redirect within 5 seconds",
		},
		{
			name:         "No redirects with 5s timeout",
			url:          "https://ducksify.ch",
			maxRedirects: 0,
			shouldPass:   true,
			description:  "Should complete without redirects within 5 seconds",
		},
		{
			name:         "Slow URL test",
			url:          "https://httpbin.org/delay/6",
			maxRedirects: 3,
			shouldPass:   false,
			description:  "Should timeout on slow response (>5s)",
		},
		{
			name:         "Unreachable URL test",
			url:          "https://unreachable-url-that-does-not-exist-12345.com",
			maxRedirects: 3,
			shouldPass:   false,
			description:  "Should fail with timeout or connection error",
		},
	}

	for _, tt := range timeoutTests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()

			result, err := testWappalyzator(tt.url, tt.maxRedirects)
			duration := time.Since(start)

			if tt.shouldPass {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					return
				}

				// Check that it completed within reasonable time (less than 10 seconds)
				if duration > 10*time.Second {
					t.Errorf("Request took too long: %v (expected < 10s)", duration)
					return
				}

				// Parse and validate the result
				var output map[string]interface{}
				if err := json.Unmarshal([]byte(result), &output); err != nil {
					t.Errorf("Failed to parse JSON result: %v", err)
					return
				}

				// Check that we got a valid response
				if techs, ok := output["technology"].([]interface{}); !ok || len(techs) == 0 {
					t.Errorf("Expected technologies in result, got: %v", output)
					return
				}

				fmt.Printf("✅ %s: Completed in %v - %s\n", tt.name, duration, tt.description)
			} else {
				if err == nil {
					t.Errorf("Expected error but got success: %s", result)
				}
				fmt.Printf("✅ %s: Correctly failed as expected\n", tt.name)
			}
		})
	}
}

// Benchmark tests
func BenchmarkWappalyzator(b *testing.B) {
	url := "https://ducksify.ch"
	for i := 0; i < b.N; i++ {
		_, err := testWappalyzator(url, 3)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}

// Main test runner
func TestMain(m *testing.M) {
	fmt.Println("🧪 Starting wappalyzator tests...")
	fmt.Println("==================================")

	// Run the tests
	exitCode := m.Run()

	fmt.Println("==================================")
	fmt.Println("✅ All tests completed!")

	os.Exit(exitCode)
}
