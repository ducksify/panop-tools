package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const userAgent = "Mozilla/5.0 (compatible; Panop/1.0; +https://panop.io/)"

type httpClient struct {
	client *http.Client
	retry  int
	log    *logger
}

func newHTTPClient(timeout time.Duration, log *logger) *httpClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // skip TLS cert verification as requested
		},
	}

	return &httpClient{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		log: log,
	}
}

func (c *httpClient) setRetry(n int) {
	if n < 0 {
		n = 0
	}
	c.retry = n
}

func (c *httpClient) fetchText(rawURL string, referer string) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retry; attempt++ {
		body, err := c.fetchOnce(rawURL, referer)
		if err == nil {
			return body, nil
		}
		lastErr = err
		c.log.Debugf("fetch %s failed (attempt %d/%d): %v", rawURL, attempt+1, c.retry+1, err)
	}

	return "", lastErr
}

func (c *httpClient) fetchOnce(rawURL string, referer string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	//req.Header.Set("Accept-Encoding", "identity")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	c.log.Debugf("GET %s", rawURL)

	resp, err := c.client.Do(req)
	if err != nil {
		// TLS errors or other network errors should be skipped silently,
		// so just return the error to be logged in debug mode only.
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	c.log.Debugf("Fetched %s (%d bytes)", parsed.Host, len(data))
	return string(data), nil
}
