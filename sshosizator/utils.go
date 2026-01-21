package main

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

func loadRulesFromBytes(data []byte) ([]compiledRule, error) {
	var rf rulesFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, err
	}

	var out []compiledRule
	for _, r := range rf.Rules {
		cr := compiledRule{
			rule:         r,
			hasEquals:    r.Match.Equals != "",
			hasContains:  r.Match.Contains != "",
			hasLoginCont: r.Match.LoginContains != "",
		}
		if r.Match.Regex != "" {
			re, err := regexp.Compile(r.Match.Regex)
			if err != nil {
				return nil, fmt.Errorf("invalid regex for rule %q: %w", r.ID, err)
			}
			cr.bannerRegex = re
		}
		if r.Match.LoginRegex != "" {
			re, err := regexp.Compile(r.Match.LoginRegex)
			if err != nil {
				return nil, fmt.Errorf("invalid login_regex for rule %q: %w", r.ID, err)
			}
			cr.loginRegex = re
		}
		out = append(out, cr)
	}
	return out, nil
}

func matchRules(rules []compiledRule, banner, login string) (ruleConfig, string, bool) {
	for _, cr := range rules {
		// Prefer login banner match if the rule is defined that way.
		if cr.hasLoginCont && login != "" && strings.Contains(login, cr.rule.Match.LoginContains) {
			return cr.rule, "login_banner", true
		}
		if cr.loginRegex != nil && login != "" && cr.loginRegex.MatchString(login) {
			return cr.rule, "login_banner", true
		}

		if cr.hasEquals && banner == cr.rule.Match.Equals {
			return cr.rule, "banner", true
		}
		if cr.hasContains && strings.Contains(banner, cr.rule.Match.Contains) {
			return cr.rule, "banner", true
		}
		if cr.bannerRegex != nil && cr.bannerRegex.MatchString(banner) {
			return cr.rule, "banner", true
		}
	}
	return ruleConfig{}, "unknown", false
}

// readBanner grabs the SSH identification banner by opening a TCP connection
// and reading the first line sent by the server.
func readBanner(addr string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Trim CRLF.
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line, nil
}

// readLoginBanner performs a minimal SSH handshake using golang.org/x/crypto/ssh
// to capture any pre-authentication banner text. It does not use real credentials
// and will typically see an auth failure; that's fine, we just care about the banner.
func readLoginBanner(host string, port int, timeout time.Duration) (string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	var bannerText string

	cfg := &ssh.ClientConfig{
		// Dummy username; servers usually show the auth banner before rejecting auth.
		User: "bannerprobe",
		Auth: []ssh.AuthMethod{
			// No real auth methods; we expect auth to fail.
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
		BannerCallback: func(message string) error {
			bannerText = message
			return nil
		},
	}

	// We ignore the returned client and most errors; the banner callback may already
	// have fired even if Dial ultimately fails.
	client, err := ssh.Dial("tcp", addr, cfg)
	if client != nil {
		_ = client.Close()
	}

	if bannerText != "" {
		return bannerText, nil
	}
	if err != nil {
		return "", err
	}
	return "", nil
}
