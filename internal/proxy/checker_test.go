package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"proxy-checker/internal/config"
	"strings"
	"testing"
	"time"
)

func TestDoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("127.0.0.1")) // используем IP 127.0.0.1
	}))
	defer server.Close()

	_, port, _ := strings.Cut(server.Listener.Addr().String(), ":")
	proxyAddress := "127.0.0.1:" + port

	cfg := config.ProxyChecker{
		API:         server.URL,
		Timeout:     5 * time.Second, // Увеличен таймаут для предотвращения timeouts
		Concurrency: 1,
	}
	checker := NewChecker(cfg)

	err := checker.(*ProxyChecker).doRequest(context.Background(), "http", proxyAddress)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Тестирование функции doRequest при несоответствии IP
func TestDoRequest_IPMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("111.111.111.111"))
	}))
	defer server.Close()

	serverURL := server.Listener.Addr().String()
	_, port, _ := strings.Cut(serverURL, ":")
	proxyAddress := "127.0.0.1:" + port

	cfg := config.ProxyChecker{
		API:         server.URL,
		Timeout:     5 * time.Second,
		Concurrency: 1,
	}
	checker := NewChecker(cfg)

	err := checker.(*ProxyChecker).doRequest(context.Background(), "http", proxyAddress)

	if err == nil || !strings.Contains(err.Error(), "proxy IP mismatch") {
		t.Fatalf("expected IP mismatch error, got %v", err)
	}
}

func TestCheckOne_ValidProxy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("127.0.0.1"))
	}))
	defer server.Close()

	serverURL := server.Listener.Addr().String()
	_, port, _ := strings.Cut(serverURL, ":")
	proxyAddress := "127.0.0.1:" + port

	cfg := config.ProxyChecker{
		API:         server.URL,
		Timeout:     5 * time.Second,
		Concurrency: 1,
	}
	checker := NewChecker(cfg)

	proxy, err := checker.CheckOne(context.Background(), proxyAddress)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if proxy != proxyAddress {
		t.Fatalf("expected proxy '%s', got %v", proxyAddress, proxy)
	}
}

func TestCheckOne_InvalidProxy(t *testing.T) {
	cfg := config.ProxyChecker{
		API:         "http://example.com",
		Timeout:     5 * time.Second,
		Concurrency: 1,
	}
	checker := NewChecker(cfg)

	_, err := checker.CheckOne(context.Background(), "invalid_proxy")

	if err == nil || !strings.Contains(err.Error(), "invalid proxy url") {
		t.Fatalf("expected invalid proxy URL error, got %v", err)
	}
}

func TestAwaitCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("127.0.0.1"))
	}))
	defer server.Close()

	serverURL := server.Listener.Addr().String()
	_, port, _ := strings.Cut(serverURL, ":")
	proxyAddress := "127.0.0.1:" + port

	cfg := config.ProxyChecker{
		API:         server.URL,
		Timeout:     5 * time.Second,
		Concurrency: 2,
	}
	checker := NewChecker(cfg)

	proxiesCh := make(chan string, 3)
	proxiesCh <- proxyAddress
	proxiesCh <- "111.111.111.111:8080"
	proxiesCh <- "invalid_proxy"
	close(proxiesCh)

	results, err := checker.AwaitCheck(context.Background(), proxiesCh)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedResults := []string{proxyAddress}
	if len(results) != len(expectedResults) || results[0] != expectedResults[0] {
		t.Fatalf("expected results %v, got %v", expectedResults, results)
	}
}
