package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

//go:embed banners.yml
var embeddedBanners []byte

func main() {
	host := pflag.String("host", "", "target host")
	port := pflag.Int("port", 22, "target port")
	timeout := pflag.Duration("timeout", 5*time.Second, "connection and read timeout")
	pflag.Parse()

	if *host == "" {
		fmt.Fprintln(os.Stderr, "missing required --host")
		os.Exit(1)
	}
	if *port <= 0 || *port > 65535 {
		fmt.Fprintln(os.Stderr, "invalid --port")
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)

	if len(embeddedBanners) == 0 {
		fmt.Fprintln(os.Stderr, "embedded banner rules are missing")
		os.Exit(1)
	}

	rules, err := loadRulesFromBytes(embeddedBanners)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load embedded banner rules: %v\n", err)
		os.Exit(1)
	}

	banner, err := readBanner(addr, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read SSH banner from %s: %v\n", addr, err)
		os.Exit(1)
	}

	// For the SSH login/auth banner we pass host+port separately, since the SSH
	// client constructs its own "host:port" address string.
	// Try to read an SSH login/auth banner via a minimal SSH handshake. Some
	// servers will not send such a banner or will fail auth immediately; in
	// that case we just continue with banner-only detection.
	var loginBanner string
	loginBanner, _ = readLoginBanner(*host, *port, *timeout)

	rule, source, ok := matchRules(rules, banner, loginBanner)
	var res result
	if ok {
		res = result{
			OS:          rule.OS,
			OSShortname: rule.OSShortname,
			Ver:         rule.Version,
			Source:      source,
			RuleID:      rule.ID,
			Banner:      banner,
			Login:       loginBanner,
		}
	} else {
		res = result{
			OS:          "Unknown",
			OSShortname: "",
			Ver:         "",
			Source:      "unknown",
			RuleID:      "",
			Banner:      banner,
			Login:       loginBanner,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}
