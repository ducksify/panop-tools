package dkim

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// Record represents a parsed DKIM record
type Record struct {
	Version   string
	PublicKey string
	TestMode  bool
	Fields    map[string]string
}

// ParseTXT parses a DKIM TXT record from DNS
func ParseTXT(txtData []string) (*Record, error) {
	// Join multi-string TXT records and remove quotes with spaces
	fullTxt := strings.Join(txtData, "")
	fullTxt = regexp.MustCompile(`"\s+"`).ReplaceAllString(fullTxt, " ")

	record := &Record{
		Fields: make(map[string]string),
	}

	// Parse key=value pairs
	re := regexp.MustCompile(`\s*([^=]+)=([^;]*?)(\\?;\s*|\s*$)`)
	matches := re.FindAllStringSubmatch(fullTxt, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		record.Fields[key] = value
	}

	// Extract common fields
	if v, ok := record.Fields["v"]; ok {
		record.Version = v
	}
	if p, ok := record.Fields["p"]; ok {
		// Remove whitespace and pad base64
		p = regexp.MustCompile(`\s+`).ReplaceAllString(p, "")
		record.PublicKey = base64Pad(p)
	}
	if t, ok := record.Fields["t"]; ok {
		record.TestMode = strings.ToLower(t) == "y"
	}

	return record, nil
}

// base64Pad pads base64 string to multiple of 4
func base64Pad(b64 string) string {
	for len(b64)%4 != 0 {
		b64 += "="
	}
	return b64
}

// DecodePublicKey decodes the base64 public key
func (r *Record) DecodePublicKey() ([]byte, error) {
	if r.PublicKey == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(r.PublicKey)
}
