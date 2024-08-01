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

	NewChecker(cfg)
}

func TestCheck(t *testing.T) {

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

	cfg := config.ProxyChecker{
		API:     server.URL,
		Timeout: 5 * time.Second,
	}
	checker := NewChecker(cfg)

	_, err = checker.Check(ctx, line)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	invalidLine := "invalid proxy"
	_, err = checker.Check(ctx, invalidLine)
	if err == nil {
		t.Errorf("expected error for invalid proxy line, got none")
	}
}
