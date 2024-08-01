package handler

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChecker is a mock implementation of proxy.Checker interface
type MockChecker struct {
	mock.Mock
}

func (m *MockChecker) Check(ctx context.Context, proxy string) (string, error) {
	args := m.Called(ctx, proxy)
	return args.String(0), args.Error(1)
}

// Test for ProxyCheckApi
func TestProxyCheckApi(t *testing.T) {
	checker := new(MockChecker)
	checker.On("Check", mock.Anything, "http://validproxy").Return("http://validproxy", nil)

	handler := ProxyCheckApi(checker)

	req := httptest.NewRequest("POST", "/api/proxy_check", strings.NewReader(`["http://validproxy"]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	checker.AssertCalled(t, "Check", mock.Anything, "http://validproxy")
}

func TestProxyCheckApi_InvalidRequest(t *testing.T) {
	handler := ProxyCheckApi(nil)

	req := httptest.NewRequest("POST", "/api/proxy_check", strings.NewReader(`[]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test for ProxyCheckForm
func TestProxyCheckForm(t *testing.T) {
	tmpl, err := template.New("proxy_check_form.html.tmpl").Parse("form template")
	assert.NoError(t, err)

	handler := ProxyCheckForm(tmpl)

	req := httptest.NewRequest("GET", "/form", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "form template", w.Body.String())
}

// Test for ProxyCheckWeb
func TestProxyCheckWeb(t *testing.T) {
	tmpl, err := template.New("proxies_table.html.tmpl").Parse("table template {{.}}")
	assert.NoError(t, err)

	checker := new(MockChecker)
	checker.On("Check", mock.Anything, "http://validproxy").Return("http://validproxy", nil)

	handler := ProxyCheckWeb(tmpl, checker)

	form := "proxies=http://validproxy"
	req := httptest.NewRequest("POST", "/web", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "http://validproxy")
}

func TestProxyCheckWeb_InvalidForm(t *testing.T) {
	tmpl, err := template.New("proxies_table.html.tmpl").Parse("table template {{.}}")
	assert.NoError(t, err)

	handler := ProxyCheckWeb(tmpl, nil)

	req := httptest.NewRequest("POST", "/web", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProxyCheckWeb_MissingProxies(t *testing.T) {
	tmpl, err := template.New("proxies_table.html.tmpl").Parse("table template {{.}}")
	assert.NoError(t, err)

	handler := ProxyCheckWeb(tmpl, nil)

	req := httptest.NewRequest("POST", "/web", strings.NewReader("proxies="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
