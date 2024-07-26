package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"proxy-checker/internal/config"
	"regexp"
	"time"
)

var pattern *regexp.Regexp

func init() {
	pattern = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b`)
}

type Checker struct {
	Target  string
	Timeout time.Duration
}

func NewChecker(cfg config.ProxyChecker) *Checker {
	return &Checker{Target: cfg.API, Timeout: cfg.Timeout}
}

func (c *Checker) Check(ctx context.Context, line string) error {
	var proxy string
	if proxy = pattern.FindString(line); proxy == "" {
		return fmt.Errorf("invalid proxy url: %s", line)
	}

	proxyURL := http.ProxyURL(&url.URL{
		Host: proxy,
	})

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: proxyURL,
		},
		Timeout: c.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.Target, nil)
	if err != nil {
		return err
	}

	slog.Debug("Doing request to", slog.String("proxy", line))
	resp, err := client.Do(req)
	slog.Debug("Request finished", slog.String("proxy", line), slog.Any("error", err))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	return nil
}
