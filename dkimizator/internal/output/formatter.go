package output

import (
	"encoding/json"
	"io"

	"github.com/ducksify/panop-tools/dkimizator/internal/crypto"
)

// Result represents a single DKIM result for JSON output
type Result struct {
	FQDN        string   `json:"fqdn"`
	TXT         []string `json:"txt"`
	Selector    string   `json:"selector"`
	Domain      string   `json:"domain"`
	Fingerprint string   `json:"fingerprint"`
	Size        int      `json:"size"`
	Modulus     string   `json:"modulus"`
	Exponent    string   `json:"exponent"`
	Mode        string   `json:"mode"`
	X509Key     string   `json:"x509_key"`
}

// Output represents the complete JSON output structure
type Output struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}

// Formatter handles output formatting
type Formatter struct {
	writer  io.Writer
	quiet   bool
	results []Result
}

// NewFormatter creates a new output formatter
func NewFormatter(writer io.Writer, quiet bool) *Formatter {
	return &Formatter{
		writer:  writer,
		quiet:   quiet,
		results: make([]Result, 0),
	}
}

// AddResult adds a result to the collection
func (f *Formatter) AddResult(fqdn string, txt []string, keyInfo *crypto.KeyInfo, domain, selector, mode string, x509Key string) {
	result := Result{
		FQDN:        fqdn,
		TXT:         txt,
		Selector:    selector,
		Domain:      domain,
		Fingerprint: keyInfo.Fingerprint,
		Size:        keyInfo.Size,
		Modulus:     keyInfo.Modulus.String(),
		Exponent:    keyInfo.Exponent.String(),
		Mode:        mode,
		X509Key:     x509Key,
	}
	f.results = append(f.results, result)
}

// OutputJSON outputs all collected results as JSON
func (f *Formatter) OutputJSON() error {
	output := Output{
		Count:   len(f.results),
		Results: f.results,
	}
	return json.NewEncoder(f.writer).Encode(output)
}
