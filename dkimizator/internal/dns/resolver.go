package dns

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/zmap/dns"
	"github.com/zmap/zdns/v2/src/zdns"
)

// QueryResult represents the result of a DNS query
type QueryResult struct {
	Selector string
	FQDN     string
	TXT      []string
	Error    error
	Found    bool
}

// QuerySelectors queries DNS for multiple selectors using zdns library with iterative resolution
func QuerySelectors(ctx context.Context, selectors []string, domain string, timeout time.Duration) <-chan *QueryResult {
	resultChan := make(chan *QueryResult, len(selectors))

	// Build FQDNs and create mapping from FQDN to selector
	fqdns := make([]string, 0, len(selectors))
	fqdnToSelector := make(map[string]string, len(selectors))
	for _, selector := range selectors {
		fqdn := fmt.Sprintf("%s._domainkey.%s", selector, domain)
		fqdns = append(fqdns, fqdn)
		fqdnToSelector[fqdn] = selector
	}

	// Process queries in a goroutine
	go func() {
		defer close(resultChan)

		if len(fqdns) == 0 {
			return
		}

		// Create a context with timeout
		queryCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Create zdns resolver configuration with iterative mode
		// This config is shared across all workers
		config := zdns.NewResolverConfig()
		// Use per-query timeout (zdns default is 15 seconds)
		// The overall operation timeout is handled by queryCtx
		config.Timeout = 15 * time.Second
		config.IterativeTimeout = 8 * time.Second // Default iterative timeout
		config.NetworkTimeout = 2 * time.Second   // Default network timeout
		config.Retries = 3                        // Default retries
		config.MaxDepth = 10                      // Default max depth

		// Create a channel to distribute FQDNs to workers
		fqdnChan := make(chan string, len(fqdns))
		for _, fqdn := range fqdns {
			fqdnChan <- fqdn
		}
		close(fqdnChan)

		// Number of concurrent workers (similar to zdns default of 100 threads)
		numWorkers := 100
		if len(fqdns) < numWorkers {
			numWorkers = len(fqdns)
		}

		// Use WaitGroup to wait for all workers to complete
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		// Start worker goroutines
		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()

				// Each worker creates its own resolver (zdns resolvers are not thread-safe)
				resolver, err := zdns.InitResolver(config)
				if err != nil {
					// If resolver creation fails, log error and return
					// Other workers will handle the FQDNs from the shared channel
					slog.Default().Error("failed to initialize zdns resolver", "error", err)
					return
				}
				defer resolver.Close()

				// Process FQDNs from the channel
				for fqdn := range fqdnChan {
					// Check for context cancellation
					select {
					case <-queryCtx.Done():
						// Context expired, send error for this FQDN and drain remaining
						selector, ok := fqdnToSelector[fqdn]
						if !ok {
							parts := strings.Split(fqdn, "._domainkey.")
							if len(parts) == 2 {
								selector = parts[0]
							} else {
								selector = fqdn
							}
						}
						resultChan <- &QueryResult{
							Selector: selector,
							FQDN:     fqdn,
							Error:    fmt.Errorf("query timeout: %w", queryCtx.Err()),
							Found:    false,
						}
						// Drain remaining FQDNs from this worker's channel
						for remainingFqdn := range fqdnChan {
							remainingSelector, ok := fqdnToSelector[remainingFqdn]
							if !ok {
								parts := strings.Split(remainingFqdn, "._domainkey.")
								if len(parts) == 2 {
									remainingSelector = parts[0]
								} else {
									remainingSelector = remainingFqdn
								}
							}
							resultChan <- &QueryResult{
								Selector: remainingSelector,
								FQDN:     remainingFqdn,
								Error:    fmt.Errorf("query timeout: %w", queryCtx.Err()),
								Found:    false,
							}
						}
						return
					default:
					}

					// Create DNS question for TXT record
					question := &zdns.Question{
						Name:  fqdn,
						Type:  dns.TypeTXT,
						Class: dns.ClassINET,
					}

					// Perform iterative lookup
					result, trace, status, lookupErr := resolver.IterativeLookup(queryCtx, question)

					// Map FQDN back to selector
					selector, ok := fqdnToSelector[fqdn]
					if !ok {
						// Try to extract selector from FQDN format
						parts := strings.Split(fqdn, "._domainkey.")
						if len(parts) == 2 {
							selector = parts[0]
						} else {
							selector = fqdn
						}
					}

					queryResult := &QueryResult{
						Selector: selector,
						FQDN:     fqdn,
					}

					// Handle lookup errors
					if lookupErr != nil {
						queryResult.Error = fmt.Errorf("dns lookup error: %w", lookupErr)
						queryResult.Found = false
						resultChan <- queryResult
						continue
					}

					// Handle status codes
					if status != zdns.StatusNoError {
						// Map status to error message
						switch status {
						case zdns.StatusNXDomain:
							queryResult.Error = fmt.Errorf("domain not found: %s", status)
						case zdns.StatusTimeout, zdns.StatusIterTimeout:
							queryResult.Error = fmt.Errorf("query timeout: %s", status)
						case zdns.StatusServFail:
							queryResult.Error = fmt.Errorf("server failure: %s", status)
						case zdns.StatusRefused:
							queryResult.Error = fmt.Errorf("query refused: %s", status)
						default:
							queryResult.Error = fmt.Errorf("dns error: %s", status)
						}
						queryResult.Found = false
						resultChan <- queryResult
						continue
					}

					// Extract TXT records from result
					if result != nil {
						var txtRecords []string
						for _, answer := range result.Answers {
							if ans, ok := answer.(zdns.Answer); ok {
								// Check if it's a TXT record
								if ans.Type == "TXT" || ans.RrType == dns.TypeTXT {
									if ans.Answer != "" {
										// Clean up the answer - remove quotes and handle multi-string TXT records
										answerText := strings.Trim(ans.Answer, "\"")
										txtRecords = append(txtRecords, answerText)
									}
								}
							}
						}

						if len(txtRecords) > 0 {
							queryResult.TXT = txtRecords
							queryResult.Found = true
						} else {
							queryResult.Found = false
						}
					} else {
						queryResult.Found = false
					}

					// Suppress unused variable warning for trace
					_ = trace

					// Send result immediately (no need to collect all results first)
					logger := slog.Default()
					if queryResult.Error != nil {
						logger.Debug("DNS query error", "fqdn", queryResult.FQDN, "error", queryResult.Error)
					} else if queryResult.Found {
						logger.Debug("DNS query found", "fqdn", queryResult.FQDN, "txt_count", len(queryResult.TXT))
					} else {
						logger.Debug("DNS query not found", "fqdn", queryResult.FQDN)
					}
					resultChan <- queryResult
				}
			}()
		}

		// Wait for all workers to complete
		wg.Wait()
	}()

	return resultChan
}
