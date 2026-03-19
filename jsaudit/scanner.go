package main

import (
	"net/url"
	"regexp"
	"strings"
	"time"
)

type scriptLibrary struct {
	Library    string `json:"library"`
	Version    string `json:"version"`
	Vulnerable bool   `json:"vulnerable"`
	CVEs       []struct {
		CVE         string `json:"cve"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
	} `json:"cves"`
}

type scriptEntry struct {
	File      string          `json:"file"`
	URL       string          `json:"url"`
	Status    string          `json:"status"`
	SizeBytes int             `json:"size_bytes"`
	FetchMS   int64           `json:"fetch_ms"`
	Libraries []scriptLibrary `json:"libraries"`
}

type vulnerabilityFinding struct {
	File        string `json:"file"`
	URL         string `json:"url"`
	Library     string `json:"library"`
	Version     string `json:"version"`
	CVE         string `json:"cve"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type summary struct {
	ScriptsFound      int   `json:"scripts_found"`
	FetchedOK         int   `json:"fetched_ok"`
	WAFBlocked        int   `json:"waf_blocked"`
	Failed            int   `json:"failed"`
	Vulnerabilities   int   `json:"vulnerabilities"`
	LibrariesDetected int   `json:"libraries_detected"`
	ScanDurationMS    int64 `json:"scan_duration_ms"`
}

type report struct {
	Scanner         string                 `json:"scanner"`
	Version         string                 `json:"version"`
	Target          string                 `json:"target"`
	Date            string                 `json:"date"`
	Summary         summary                `json:"summary"`
	Scripts         []scriptEntry          `json:"scripts"`
	Vulnerabilities []vulnerabilityFinding `json:"vulnerabilities"`
}

func runScan(opts cliOptions, client *httpClient, db *db, log *logger) (*report, error) {
	start := time.Now()

	client.setRetry(opts.retry)

	rep := &report{
		Scanner: "jsaudit-go",
		Version: "1.0.0",
		Target:  opts.targetURL,
		Date:    time.Now().UTC().Format(time.RFC3339),
	}

	// Fetch main page
	html, err := client.fetchText(opts.targetURL, "")
	if err != nil {
		return nil, err
	}

	scriptURLs := extractScriptURLs(html, opts.targetURL)
	rep.Summary.ScriptsFound = len(scriptURLs)

	if len(scriptURLs) == 0 {
		log.Debugf("No external <script> URLs found; only main page fetched: %s", opts.targetURL)
	} else {
		log.Debugf("Discovered %d script URLs:", len(scriptURLs))
		for i, u := range scriptURLs {
			log.Debugf("  [%d] %s", i+1, u)
		}
	}

	var scripts []scriptEntry
	var findings []vulnerabilityFinding

	for _, scriptURL := range scriptURLs {
		if opts.delay > 0 {
			time.Sleep(opts.delay)
		}

		entry := scriptEntry{
			URL:    scriptURL,
			Status: "unknown",
		}

		if u, err := url.Parse(scriptURL); err == nil {
			parts := strings.Split(u.Path, "/")
			if len(parts) > 0 {
				entry.File = parts[len(parts)-1]
			}
		}

		t0 := time.Now()
		body, err := client.fetchText(scriptURL, opts.targetURL)
		elapsed := time.Since(t0)
		entry.FetchMS = elapsed.Milliseconds()

		if err != nil {
			entry.Status = "failed"
			rep.Summary.Failed++
			scripts = append(scripts, entry)
			continue
		}

		entry.SizeBytes = len(body)

		if len(body) < 512 {
			entry.Status = "blocked"
			rep.Summary.WAFBlocked++
			scripts = append(scripts, entry)
			continue
		}

		entry.Status = "ok"
		rep.Summary.FetchedOK++

		libs := detectLibraries(scriptURL, body, db, log)
		entry.Libraries = libs

		for _, lib := range libs {
			for _, c := range lib.CVEs {
				findings = append(findings, vulnerabilityFinding{
					File:        entry.File,
					URL:         entry.URL,
					Library:     lib.Library,
					Version:     lib.Version,
					CVE:         c.CVE,
					Severity:    c.Severity,
					Description: c.Description,
				})
			}
		}

		scripts = append(scripts, entry)
	}

	rep.Scripts = scripts
	rep.Vulnerabilities = findings
	rep.Summary.Vulnerabilities = len(findings)
	totalLibs := 0
	for _, s := range scripts {
		totalLibs += len(s.Libraries)
	}
	rep.Summary.LibrariesDetected = totalLibs
	rep.Summary.ScanDurationMS = time.Since(start).Milliseconds()

	return rep, nil
}

func extractScriptURLs(html string, baseURL string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	scriptTagRe := regexp.MustCompile(`(?is)<script\b[^>]*\bsrc=["']([^"']+)["'][^>]*>`)
	linkTagRe := regexp.MustCompile(`(?is)<link\b([^>]*)>`)
	attrRe := regexp.MustCompile(`(?is)\b([a-zA-Z_:][\w:.-]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>` + "`" + `]+))`)

	seen := make(map[string]struct{})
	var urls []string

	for _, m := range scriptTagRe.FindAllStringSubmatch(html, -1) {
		if len(m) < 2 {
			continue
		}

		addResolvedURL(strings.TrimSpace(m[1]), base, seen, &urls)
	}

	for _, m := range linkTagRe.FindAllStringSubmatch(html, -1) {
		if len(m) < 2 {
			continue
		}

		attrs := m[1]
		rel := ""
		as := ""
		href := ""

		for _, am := range attrRe.FindAllStringSubmatch(attrs, -1) {
			if len(am) < 5 {
				continue
			}

			val := am[2]
			if val == "" {
				val = am[3]
			}
			if val == "" {
				val = am[4]
			}

			switch strings.ToLower(am[1]) {
			case "rel":
				rel = strings.ToLower(strings.TrimSpace(val))
			case "as":
				as = strings.ToLower(strings.TrimSpace(val))
			case "href":
				href = strings.TrimSpace(val)
			}
		}

		if href == "" {
			continue
		}

		// Include JS files hinted via module/preload links.
		if !(hasRelToken(rel, "modulepreload") || (hasRelToken(rel, "preload") && as == "script")) {
			continue
		}

		addResolvedURL(href, base, seen, &urls)
	}

	return urls
}

func hasRelToken(rel, token string) bool {
	for _, part := range strings.Fields(strings.ToLower(rel)) {
		if part == token {
			return true
		}
	}
	return false
}

func addResolvedURL(raw string, base *url.URL, seen map[string]struct{}, urls *[]string) {
	if raw == "" {
		return
	}

	u, err := url.Parse(raw)
	if err != nil {
		return
	}
	if !u.IsAbs() {
		u = base.ResolveReference(u)
	}

	final := u.String()
	if _, ok := seen[final]; ok {
		return
	}
	seen[final] = struct{}{}
	*urls = append(*urls, final)
}

func detectLibraries(scriptURL, content string, db *db, log *logger) []scriptLibrary {
	found := make(map[string]string) // lib -> version

	for lib, patterns := range db.URLPatterns {
		for _, pat := range patterns {
			reStr := strings.ReplaceAll(pat, "§§version§§", `(\\d+[\\d.]*)`)
			re, err := regexp.Compile("(?i)" + reStr)
			if err != nil {
				continue
			}
			if m := re.FindStringSubmatch(scriptURL); len(m) > 1 {
				found[lib] = m[1]
				break
			}
		}
	}

	for lib, patterns := range db.ContentRegex {
		for _, pat := range patterns {
			reStr := strings.ReplaceAll(pat, "§§version§§", `(\\d+[\\d.]*)`)
			re, err := regexp.Compile("(?i)" + reStr)
			if err != nil {
				continue
			}
			if m := re.FindStringSubmatch(content); len(m) > 1 {
				if _, exists := found[lib]; !exists {
					found[lib] = m[1]
				}
				break
			}
		}
	}

	// Cross-check detected versions against canonical header comment version.
	// If the header declares a higher version, prefer it (to avoid picking up
	// requirement strings like ">= 2.6.0" as the actual library version).
	if headerVersion := extractHeaderVersion(content); headerVersion != "" {
		for lib, version := range found {
			if compareVersions(headerVersion, version) > 0 {
				log.Debugf("Version override for %s: %s -> %s (header takes precedence)", lib, version, headerVersion)
				found[lib] = headerVersion
			}
		}
	}

	var libs []scriptLibrary

	for lib, version := range found {
		entry, ok := db.Libs[lib]
		if !ok {
			continue
		}

		var cves []struct {
			CVE         string `json:"cve"`
			Severity    string `json:"severity"`
			Description string `json:"description"`
		}
		for _, v := range entry.Vulns {
			if !versionInRange(version, v.AtOrAbove, v.Below) {
				continue
			}
			cves = append(cves, struct {
				CVE         string `json:"cve"`
				Severity    string `json:"severity"`
				Description string `json:"description"`
			}{
				CVE:         v.CVE,
				Severity:    strings.ToUpper(v.Severity),
				Description: v.Info,
			})
		}

		libs = append(libs, scriptLibrary{
			Library:    entry.Name,
			Version:    version,
			Vulnerable: len(cves) > 0,
			CVEs:       cves,
		})
	}

	return libs
}

func versionInRange(version, atOrAbove, below string) bool {
	if atOrAbove != "" && compareVersions(version, atOrAbove) < 0 {
		return false
	}
	if below != "" && compareVersions(version, below) >= 0 {
		return false
	}
	return true
}

func compareVersions(a, b string) int {
	parse := func(s string) []int {
		parts := strings.Split(s, ".")
		out := make([]int, len(parts))
		for i, p := range parts {
			n := 0
			for _, ch := range p {
				if ch < '0' || ch > '9' {
					break
				}
				n = n*10 + int(ch-'0')
			}
			out[i] = n
		}
		return out
	}

	pa := parse(a)
	pb := parse(b)

	max := len(pa)
	if len(pb) > max {
		max = len(pb)
	}

	for i := 0; i < max; i++ {
		va := 0
		vb := 0
		if i < len(pa) {
			va = pa[i]
		}
		if i < len(pb) {
			vb = pb[i]
		}
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
	}
	return 0
}

// Header version patterns mirror HEADER_VERSION_PATTERNS from the Node.js
// implementation. These are used to extract a canonical version from header
// comments (e.g. "//! version : 2.30.1") to avoid false positives.
var headerVersionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)//!\s*version\s*:\s*(\d+\.\d+[\d.]*)`),
	regexp.MustCompile(`(?i)\*\s*@version\s+(\d+\.\d+[\d.]*)`),
	regexp.MustCompile(`(?i)/\*![\s\S]{0,300}?version\s*:\s*(\d+\.\d+[\d.]*)`),
}

func extractHeaderVersion(content string) string {
	for _, re := range headerVersionPatterns {
		if m := re.FindStringSubmatch(content); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}
