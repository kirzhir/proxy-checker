package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"proxy-checker/internal/config"
	"regexp"
	"strings"
	"time"
)

var pattern *regexp.Regexp

type Checker interface {
	Check(ctx context.Context, line string) (string, error)
}

func init() {
	pattern = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b`)
}

type ProxyChecker struct {
	Target  string
	Timeout time.Duration
}

func NewChecker(cfg config.ProxyChecker) Checker {
	return &ProxyChecker{Target: cfg.API, Timeout: cfg.Timeout}
}

func (c *ProxyChecker) Check(ctx context.Context, line string) (string, error) {
	var proxy string
	if proxy = pattern.FindString(line); proxy == "" {
		return proxy, fmt.Errorf("invalid proxy url: %s", line)
	}

	r := make(chan error)

	go func() {
		r <- c.doRequest(ctx, "http", proxy)
	}()

	go func() {
		r <- c.doRequest(ctx, "socks5", proxy)
	}()

	var err error
	for i := 0; i < 2; i++ {
		if err = <-r; err == nil {
			return proxy, nil
		}
	}

	return proxy, err
}

func (c *ProxyChecker) doRequest(ctx context.Context, schema, proxy string) error {

	proxyURL := http.ProxyURL(&url.URL{
		Host:   proxy,
		Scheme: schema,
	})

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: proxyURL,
		},
		Timeout: c.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Target, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if !strings.Contains(string(body), strings.Split(proxy, ":")[0]) {
		return fmt.Errorf("proxy IP mismatch: %s", proxy)
	}

	return nil
}
