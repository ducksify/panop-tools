package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	wappalyzer "github.com/ducksify/wappalyzergo"
)

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

	resp, err := http.DefaultClient.Get(url)
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
