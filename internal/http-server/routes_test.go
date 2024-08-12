package http_server

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"proxy-checker/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{}

	tmpl := template.New("")

	handler := New(cfg, tmpl)

	assert.NotNil(t, handler, "handler should not be nil")
}

func TestHandleHealthz(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	handler := New(&config.Config{}, template.New(""))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "I'm alive", rr.Body.String())
}

func TestHandleIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/ip", nil)
	rr := httptest.NewRecorder()

	handler := New(&config.Config{}, template.New(""))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), req.RemoteAddr)
}
