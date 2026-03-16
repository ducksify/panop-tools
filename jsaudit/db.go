package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed jsaudit-db.json
var embeddedRetireDB []byte

const retireDBURL = "https://raw.githubusercontent.com/RetireJS/retire.js/master/repository/jsrepository-v5.json"

type vuln struct {
	Below     string `json:"below,omitempty"`
	AtOrAbove string `json:"atOrAbove,omitempty"`
	CVE       string `json:"cve"`
	Severity  string `json:"severity"`
	Info      string `json:"info,omitempty"`
}

type libEntry struct {
	Name  string `json:"name"`
	Vulns []vuln `json:"vulns"`
}

type db struct {
	Libs          map[string]libEntry   `json:"libs"`
	URLPatterns   map[string][]string   `json:"urlPatterns"`
	ContentRegex  map[string][]string   `json:"contentPatterns"`
}

func loadDB(log *logger) (*db, error) {
	data := embeddedRetireDB

	if len(data) == 0 {
		return nil, errors.New("embedded DB is empty")
	}

	// The embedded file is the raw RetireJS DB; convert to our internal format.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse embedded RetireJS DB: %w", err)
	}
	return convertRetireDB(raw, log)
}

func refreshDB(log *logger, timeout time.Duration) error {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(http.MethodGet, retireDBURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Panop/1.0; +https://panop.io/)")

	log.Debugf("Refreshing DB from %s", retireDBURL)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("decode RetireJS DB: %w", err)
	}

	// We store the raw RetireJS DB, same format as upstream, so the embedded
	// file and on-disk file share the same schema.
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)
	path := filepath.Join(dir, "jsaudit-db.json")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(raw); err != nil {
		return err
	}

	log.Debugf("DB refreshed and written to %s", path)
	return nil
}

func convertRetireDB(raw map[string]any, log *logger) (*db, error) {
	result := &db{
		Libs:         make(map[string]libEntry),
		URLPatterns:  make(map[string][]string),
		ContentRegex: make(map[string][]string),
	}

	for key, v := range raw {
		entryMap, ok := v.(map[string]any)
		if !ok {
			continue
		}

		name := key
		if bower, ok := entryMap["bowername"].([]any); ok && len(bower) > 0 {
			if s, ok := bower[0].(string); ok && s != "" {
				name = s
			}
		}
		if npm, ok := entryMap["npmname"].(string); ok && npm != "" {
			name = npm
		}

		var vulns []vuln
		if vs, ok := entryMap["vulnerabilities"].([]any); ok {
			for _, vv := range vs {
				vm, ok := vv.(map[string]any)
				if !ok {
					continue
				}
				id, _ := vm["identifiers"].(map[string]any)
				cvesAny, _ := id["CVE"].([]any)
				if len(cvesAny) == 0 {
					continue
				}
				summary, _ := id["summary"].(string)

				below, _ := vm["below"].(string)
				atOrAbove, _ := vm["atOrAbove"].(string)
				severity, _ := vm["severity"].(string)
				if severity == "" {
					severity = "MEDIUM"
				}

				for _, c := range cvesAny {
					cve, ok := c.(string)
					if !ok || cve == "" {
						continue
					}
					vulns = append(vulns, vuln{
						Below:     below,
						AtOrAbove: atOrAbove,
						CVE:       cve,
						Severity:  severity,
						Info:      summary,
					})
				}
			}
		}

		result.Libs[key] = libEntry{
			Name:  name,
			Vulns: vulns,
		}

		extractors, _ := entryMap["extractors"].(map[string]any)
		if extractors != nil {
			if filenames, ok := extractors["filename"].([]any); ok {
				for _, p := range filenames {
					if s, ok := p.(string); ok && s != "" {
						result.URLPatterns[key] = append(result.URLPatterns[key], s)
					}
				}
			}
			if uris, ok := extractors["uri"].([]any); ok {
				for _, p := range uris {
					if s, ok := p.(string); ok && s != "" {
						result.URLPatterns[key] = append(result.URLPatterns[key], s)
					}
				}
			}
			if contents, ok := extractors["filecontent"].([]any); ok {
				for _, p := range contents {
					if s, ok := p.(string); ok && s != "" {
						result.ContentRegex[key] = append(result.ContentRegex[key], s)
					}
				}
			}
		}
	}

	// Add built-in URL/content patterns to match the Node.js implementation.
	mergeBuiltinPatterns(result)

	log.Debugf("Loaded %d libraries from RetireJS DB (with builtin patterns)", len(result.Libs))
	return result, nil
}

