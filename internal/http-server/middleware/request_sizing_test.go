package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestSizing(t *testing.T) {
	handler := RequestSizing(5*1024*1024, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		bodySize       int
		expectedStatus int
	}{
		{"SmallRequest", 1024, http.StatusOK},
		{"MaxSizeRequest", 5 * 1024 * 1024, http.StatusOK},
		{"TooLargeRequest", 5*1024*1024 + 1, http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewReader(make([]byte, tt.bodySize))
			req := httptest.NewRequest(http.MethodPost, "/", body)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if status := rr.Result().StatusCode; status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}
