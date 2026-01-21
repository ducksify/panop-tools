package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	product := pflag.String("product", "", "product name (e.g., ubuntu, debian)")
	version := pflag.String("version", "", "product version (e.g., 20.04, 12)")
	pflag.Parse()

	if *product == "" {
		fmt.Fprintln(os.Stderr, "missing required --product")
		os.Exit(1)
	}
	if *version == "" {
		fmt.Fprintln(os.Stderr, "missing required --version")
		os.Exit(1)
	}

	normalizedVersion := normalizeVersion(*version)
	apiURL := fmt.Sprintf("https://endoflife.date/api/%s.json", *product)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch EOL data: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "API returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read response: %v\n", err)
		os.Exit(1)
	}

	var cycles []cycleInfo
	if err := json.Unmarshal(body, &cycles); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse API response: %v\n", err)
		os.Exit(1)
	}

	if len(cycles) == 0 {
		fmt.Fprintf(os.Stderr, "no EOL data available for product %s\n", *product)
		os.Exit(1)
	}

	var latestVersion string
	if len(cycles) > 0 && cycles[0].Latest != "" {
		latestVersion = cycles[0].Latest
	}

	var matchedCycle *cycleInfo
	for i := range cycles {
		if cycles[i].Cycle == normalizedVersion {
			matchedCycle = &cycles[i]
			break
		}
	}

	currentDate := time.Now()
	var result output

	if matchedCycle == nil {
		result = output{
			ProductName:    *product,
			ProductVersion: *version,
			LatestVersion:  latestVersion,
			Supported:      false,
			ReleaseDate:    "unknown",
			EOLDate:        "unknown",
		}
	} else {
		eolDateStr := "unknown"
		if matchedCycle.EOL != false {
			if eolStr, ok := matchedCycle.EOL.(string); ok && eolStr != "" {
				eolDateStr = eolStr
			}
		}

		result = output{
			ProductName:    *product,
			ProductVersion: *version,
			LatestVersion:  latestVersion,
			Supported:      isSupported(matchedCycle.EOL, currentDate),
			ReleaseDate:    matchedCycle.ReleaseDate,
			EOLDate:        eolDateStr,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}
