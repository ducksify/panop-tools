package main

type cycleInfo struct {
	Cycle       string      `json:"cycle"`
	ReleaseDate string      `json:"releaseDate"`
	EOL         interface{} `json:"eol"` // Can be string (date) or false
	Latest      string      `json:"latest,omitempty"`
}

type output struct {
	ProductName    string `json:"product_name"`
	ProductVersion string `json:"product_version"`
	LatestVersion  string `json:"latest_version"`
	Supported      bool   `json:"supported"`
	ReleaseDate    string `json:"release_date"`
	EOLDate        string `json:"eol_date"`
}
