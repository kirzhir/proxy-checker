package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"proxy-checker/internal/config"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	cfg := config.ProxyChecker{
		API:     "http://example.com",
		Timeout: 5 * time.Second,
	}
	checker := NewChecker(cfg)

	if checker.Target != cfg.API {
		t.Errorf("expected %s, got %s", cfg.API, checker.Target)
	}
	if checker.Timeout != cfg.Timeout {
		t.Errorf("expected %v, got %v", cfg.Timeout, checker.Timeout)
	}
}

func TestCheck(t *testing.T) {
	cfg := config.ProxyChecker{
		API:     "http://example.com",
		Timeout: 5 * time.Second,
	}
	checker := NewChecker(cfg)

	line := "127.0.0.1:8080"
	ctx := context.Background()

	l, err := net.Listen("http", line)
	if err != nil {
		return
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("127.0.0.1"))
	}))

	server.Listener.Close()
	server.Listener = l

	defer server.Close()

	checker.Target = server.URL

	err = checker.Check(ctx, line)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	invalidLine := "invalid proxy"
	err = checker.Check(ctx, invalidLine)
	if err == nil {
		t.Errorf("expected error for invalid proxy line, got none")
	}
}

func TestDoRequest(t *testing.T) {
	cfg := config.ProxyChecker{
		API:     "http://example.com",
		Timeout: 5 * time.Second,
	}
	checker := NewChecker(cfg)

	proxy := "127.0.0.1:8081"

	l, err := net.Listen("socks5", proxy)
	if err != nil {
		return
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("127.0.0.1"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	server.Listener.Close()
	server.Listener = l

	defer server.Close()

	checker.Target = server.URL
	ctx := context.Background()

	err = checker.doRequest(ctx, "socks5", proxy)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	invalidProxy := "invalid proxy"
	err = checker.doRequest(ctx, "http", invalidProxy)
	if err == nil {
		t.Errorf("expected error for invalid proxy, got none")
	}
}
