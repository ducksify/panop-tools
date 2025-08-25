package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	wappalyzer "github.com/ducksify/wappalyzergo"
)

var maxRedirects = 3

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: wappalyzator <url>")
		fmt.Println("Example: wappalyzator https://www.example.com")
		os.Exit(1)
	}

	url := os.Args[1]

	// Add http:// prefix if not present
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// custom http client with timeout, insecure skip verify and max redirects
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
		log.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		log.Fatal(err)
	}

	fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)

	// Convert fingerprints map to slice of technologies
	var technologies []string
	for tech := range fingerprints {
		technologies = append(technologies, tech)
	}

	// Create JSON output
	output := map[string]interface{}{
		"technology": technologies,
	}

	// Output as JSON
	jsonOutput, err := json.Marshal(output)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonOutput))
}
