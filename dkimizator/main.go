package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ducksify/panop-tools/dkimizator/internal/crypto"
	"github.com/ducksify/panop-tools/dkimizator/internal/dkim"
	"github.com/ducksify/panop-tools/dkimizator/internal/dns"
	"github.com/ducksify/panop-tools/dkimizator/internal/generator"
	"github.com/ducksify/panop-tools/dkimizator/internal/output"
	"github.com/ducksify/panop-tools/dkimizator/internal/rules"
)

func main() {
	// Set up viper
	viper.SetEnvPrefix("DKIMIZATOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("log-level", "warn")
	viper.SetDefault("timeout", 60*time.Second)
	viper.SetDefault("quiet", false)

	// Bind flags
	pflag.String("domain", "", "Domain to scan for DKIM selectors (required)")
	pflag.String("rules", "", "Path to rules file or URL (required)")
	pflag.Bool("quiet", false, "Quiet mode (minimal output)")
	pflag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pflag.Duration("timeout", 60*time.Second, "DNS query timeout")

	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()

	// Set up logging
	logLevel := viper.GetString("log-level")
	level := parseLogLevel(logLevel)
	opts := &slog.HandlerOptions{
		Level: level,
	}
	var handler slog.Handler = slog.NewTextHandler(os.Stderr, opts)
	if viper.GetBool("quiet") {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Get configuration values
	domain := viper.GetString("domain")
	rulesFile := viper.GetString("rules")
	quiet := viper.GetBool("quiet")
	timeout := viper.GetDuration("timeout")

	// Validate required flags
	if domain == "" {
		slog.Error("domain is required (use --domain flag or DKIMIZATOR_DOMAIN env var)")
		pflag.Usage()
		os.Exit(1)
	}
	if rulesFile == "" {
		slog.Error("rules is required (use --rules flag or DKIMIZATOR_RULES env var)")
		pflag.Usage()
		os.Exit(1)
	}

	// Load rules
	loader := rules.NewLoader()
	ruleList, err := loader.LoadRules(rulesFile)
	if err != nil {
		slog.Error("failed to load rules", "error", err)
		os.Exit(1)
	}

	slog.Info("loaded rules", "count", len(ruleList))

	// Generate selectors
	selectors := make(map[string]bool)
	var mu sync.Mutex

	for _, rule := range ruleList {
		err := generator.GenerateSelectors(rule, domain, func(selector string) {
			mu.Lock()
			if !selectors[selector] {
				selectors[selector] = true
				slog.Debug("generated selector", "selector", selector, "rule", rule)
			}
			mu.Unlock()
		})
		if err != nil {
			slog.Warn("failed to generate selectors from rule", "rule", rule, "error", err)
		}
	}

	selectorList := make([]string, 0, len(selectors))
	for sel := range selectors {
		selectorList = append(selectorList, sel)
	}

	slog.Info("generated selectors", "count", len(selectorList))

	// Create output formatter
	formatter := output.NewFormatter(os.Stdout, quiet)

	// Track found selectors to avoid duplicates
	foundSelectors := make(map[string]bool)
	var foundMu sync.Mutex

	// Query DNS
	ctx := context.Background()
	resultChan := dns.QuerySelectors(ctx, selectorList, domain, timeout)

	// Process results
	for result := range resultChan {
		if result.Error != nil {
			if !quiet {
				slog.Debug("DNS query error", "selector", result.Selector, "error", result.Error)
			}
			continue
		}

		if !result.Found {
			continue
		}

		// Check for duplicate
		foundMu.Lock()
		if foundSelectors[result.Selector] {
			foundMu.Unlock()
			continue
		}
		foundSelectors[result.Selector] = true
		foundMu.Unlock()

		// Parse DKIM record
		record, err := dkim.ParseTXT(result.TXT)
		if err != nil {
			slog.Debug("failed to parse DKIM record", "selector", result.Selector, "error", err)
			continue
		}

		// Check if record has public key
		if record.PublicKey == "" {
			continue
		}

		// Analyze key
		keyInfo, err := crypto.AnalyzeKey(record)
		if err != nil {
			slog.Debug("failed to analyze key", "selector", result.Selector, "error", err)
			continue
		}

		// Get key bytes for X.509 formatting
		keyBytes, err := crypto.GetKeyBytes(record)
		if err != nil {
			slog.Debug("failed to get key bytes", "selector", result.Selector, "error", err)
			continue
		}

		x509Key := crypto.FormatX509(keyBytes)

		// Add result to collection
		formatter.AddResult(
			result.FQDN,
			result.TXT,
			keyInfo,
			domain,
			result.Selector,
			keyInfo.Mode,
			x509Key,
		)
	}

	// Output all results as JSON
	if err := formatter.OutputJSON(); err != nil {
		slog.Error("failed to output JSON", "error", err)
		os.Exit(1)
	}

	slog.Info("scan complete", "found", len(foundSelectors))
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
