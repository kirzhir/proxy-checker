package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiting_AllowRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rlMiddleware := RateLimiting(time.Second * 5)(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	rr := httptest.NewRecorder()

	rlMiddleware.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}
}

func TestRateLimiting_TooManyRequests(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rlMiddleware := RateLimiting(time.Second * 5)(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	rr := httptest.NewRecorder()

	rlMiddleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	rr = httptest.NewRecorder()
	rlMiddleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("expected status code %d, got %d", http.StatusTooManyRequests, status)
	}

	expectedBody := "too many request\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestRateLimiting_ClientCleanup(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cleanUpTimeout := time.Second * 3

	rlMiddleware := RateLimiting(cleanUpTimeout)(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	rr := httptest.NewRecorder()

	rlMiddleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}

	time.Sleep(cleanUpTimeout + time.Millisecond)

	rr = httptest.NewRecorder()
	rlMiddleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, status)
	}
}
