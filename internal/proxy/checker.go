package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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

func NewChecker(target string, timeout time.Duration) *Checker {
	return &Checker{Target: target, Timeout: timeout}
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

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	return nil
}
