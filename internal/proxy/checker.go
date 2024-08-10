package proxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"proxy-checker/internal/config"
	"regexp"
	"strings"
	"sync"
	"time"
)

var pattern = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b`)

type Checker interface {
	CheckOne(ctx context.Context, line string) (string, error)
	Check(ctx context.Context, proxies <-chan string) (<-chan string, <-chan error)
	AwaitCheck(ctx context.Context, proxiesCh <-chan string) ([]string, error)
}

type DefaultChecker struct {
	Target      string
	Timeout     time.Duration
	Concurrency int
}

func NewChecker(cfg config.ProxyChecker) Checker {
	return &DefaultChecker{
		Target:      cfg.API,
		Timeout:     cfg.Timeout,
		Concurrency: cfg.Concurrency,
	}
}

func (c *DefaultChecker) AwaitCheck(ctx context.Context, proxiesCh <-chan string) ([]string, error) {
	var err error
	res := make([]string, 0, len(proxiesCh))

	resCh, errCh := c.Check(ctx, proxiesCh)

	for {
		select {
		case proxy, ok := <-resCh:
			if !ok {
				return res, err
			}
			res = append(res, proxy)
		case err = <-errCh:
			return res, err
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}
}

func (c *DefaultChecker) Check(ctx context.Context, proxiesCh <-chan string) (<-chan string, <-chan error) {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	resCh := make(chan string, c.Concurrency)

	for i := 0; i < c.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case ch, ok := <-proxiesCh:
					if !ok {
						return
					}

					if p, err := c.CheckOne(ctx, ch); err == nil {
						resCh <- p
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resCh)
		close(errCh)
	}()

	return resCh, errCh
}

func (c *DefaultChecker) CheckOne(ctx context.Context, line string) (string, error) {
	var proxy string
	if proxy = pattern.FindString(line); proxy == "" {
		return proxy, fmt.Errorf("invalid proxy url: %s", line)
	}

	r := make(chan error)

	fn := func(schema string) {
		now := time.Now()
		log := slog.With(slog.String("schema", schema), slog.String("proxy", proxy))

		log.Debug("start proxy checking")
		err := c.doRequest(ctx, schema, proxy)
		log.Debug("proxy checking finished",
			slog.String("error", errToStr(err)),
			slog.String("duration", time.Since(now).String()),
		)

		r <- err
	}

	go fn("http")
	go fn("socks5")

	var err error
	for i := 0; i < 2; i++ {
		if err = <-r; err == nil {
			return proxy, nil
		}
	}

	return proxy, err
}

func (c *DefaultChecker) doRequest(ctx context.Context, schema, proxy string) error {
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
		return fmt.Errorf("failed to create request: %w", err)
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
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if !strings.Contains(string(body), strings.Split(proxy, ":")[0]) {
		return fmt.Errorf("proxy IP mismatch: %s", proxy)
	}

	return nil
}

func errToStr(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
