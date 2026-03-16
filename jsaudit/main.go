package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

type cliOptions struct {
	targetURL string
	timeout   time.Duration
	retry     int
	delay     time.Duration
	debug     bool
	refreshDB bool
}

func parseCLI() cliOptions {
	fs := pflag.NewFlagSet("jsaudit", pflag.ExitOnError)

	timeoutSec := fs.Int("timeout", 2, "HTTP timeout in seconds")
	retry := fs.Int("retry", 2, "number of retries on network errors/timeouts")
	delayMs := fs.Int("delay", 0, "delay in milliseconds between script requests")
	debug := fs.Bool("debug", false, "enable debug logging to stderr")
	refreshDB := fs.Bool("refresh-db", false, "refresh local jsaudit-db.json from RetireJS and exit")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: jsaudit <url> [options]\n\n")
		fmt.Fprintln(fs.Output(), "Options:")
		fmt.Fprintln(fs.Output(), "  --timeout=<s>     HTTP timeout in seconds (default: 2)")
		fmt.Fprintln(fs.Output(), "  --retry=<n>       retries on network errors/timeouts (default: 2)")
		fmt.Fprintln(fs.Output(), "  --delay=<ms>      delay in milliseconds between script requests (default: 0)")
		fmt.Fprintln(fs.Output(), "  --debug           enable debug logging to stderr")
		fmt.Fprintln(fs.Output(), "  --refresh-db      refresh local jsaudit-db.json from RetireJS and exit")
	}

	// Parse flags first
	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	args := fs.Args()
	var target string
	if len(args) > 0 {
		target = args[0]
	}

	if target == "" && !*refreshDB {
		fs.Usage()
		os.Exit(1)
	}

	return cliOptions{
		targetURL: target,
		timeout:   time.Duration(*timeoutSec) * time.Second,
		retry:     *retry,
		delay:     time.Duration(*delayMs) * time.Millisecond,
		debug:     *debug,
		refreshDB: *refreshDB,
	}
}

func main() {
	opts := parseCLI()

	logger := newLogger(opts.debug)

	if opts.refreshDB {
		if err := refreshDB(logger, opts.timeout); err != nil {
			fmt.Fprintf(os.Stderr, "failed to refresh DB: %v\n", err)
			os.Exit(1)
		}
		return
	}

	db, err := loadDB(logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load vulnerability DB: %v\n", err)
		os.Exit(1)
	}

	client := newHTTPClient(opts.timeout, logger)

	report, err := runScan(opts, client, db, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

