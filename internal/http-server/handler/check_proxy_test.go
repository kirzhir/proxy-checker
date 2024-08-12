package handler

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Checker
type MockChecker struct {
	mock.Mock
}

func (m *MockChecker) CheckOne(ctx context.Context, line string) (string, error) {
	args := m.Called(ctx, line)
	return args.String(0), args.Error(1)
}

func (m *MockChecker) Check(ctx context.Context, proxies <-chan string) (<-chan string, <-chan error) {
	args := m.Called(ctx, proxies)
	return args.Get(0).(<-chan string), args.Get(1).(<-chan error)
}

func (m *MockChecker) AwaitCheck(ctx context.Context, proxiesCh <-chan string) ([]string, error) {
	args := m.Called(ctx, proxiesCh)
	return args.Get(0).([]string), args.Error(1)
}

func TestProxyRequest_Validate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		req      ProxyRequest
		expected map[string]error
	}{
		{
			name:     "empty request",
			req:      ProxyRequest{},
			expected: map[string]error{"proxies": errEmptyRequest},
		},
		{
			name: "too many proxies",
			req:  make(ProxyRequest, 101),
			expected: map[string]error{
				"proxies": errTooManyProxies,
			},
		},
		{
			name:     "valid request",
			req:      ProxyRequest{"192.168.0.1:8080"},
			expected: map[string]error{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.req.Validate(ctx)
			assert.Equal(t, tt.expected, errors)
		})
	}
}

func TestProxyCheckAPI(t *testing.T) {
	mockChecker := new(MockChecker)
	handlerFunc := ProxyCheckAPI(mockChecker)

	t.Run("Invalid JSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/check", strings.NewReader("invalid-json"))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to decode request")
	})

	t.Run("Invalid Proxies", func(t *testing.T) {
		requestBody := `[]`
		req, err := http.NewRequest("POST", "/api/check", strings.NewReader(requestBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request")
	})

	t.Run("Checking failed", func(t *testing.T) {
		mockChecker = new(MockChecker)
		handlerFunc = ProxyCheckAPI(mockChecker)
		requestBody := `["192.168.0.1:8080"]`
		req, err := http.NewRequest("POST", "/api/check", strings.NewReader(requestBody))
		assert.NoError(t, err)

		mockChecker.On("AwaitCheck", mock.Anything, mock.Anything).Return([]string{}, errors.New("some error API"))

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "{\"message\":\"Proxy check failed: some error API\",\"details\":{}}\n", rr.Body.String())
		mockChecker.AssertExpectations(t)
	})

	t.Run("Valid Proxies", func(t *testing.T) {
		mockChecker = new(MockChecker)
		handlerFunc = ProxyCheckAPI(mockChecker)
		requestBody := `["192.168.0.1:8080"]`
		req, err := http.NewRequest("POST", "/api/check", strings.NewReader(requestBody))
		assert.NoError(t, err)

		mockChecker.On("AwaitCheck", mock.Anything, mock.Anything).Return([]string{"192.168.0.1:8080"}, nil)

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockChecker.AssertExpectations(t)
	})
}

func TestProxyCheckWeb(t *testing.T) {
	tmpl := template.Must(template.New("proxies_table.html.tmpl").Parse("{{.}}"))

	t.Run("Missing proxies parameter", func(t *testing.T) {
		mockChecker := new(MockChecker)
		handlerFunc := ProxyCheckWeb(tmpl, mockChecker)
		req, err := http.NewRequest("POST", "/web/check", strings.NewReader("proxies="))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing 'proxies' parameter")
	})

	t.Run("Form parsing error", func(t *testing.T) {
		mockChecker := new(MockChecker)
		handlerFunc := ProxyCheckWeb(tmpl, mockChecker)
		req, err := http.NewRequest("POST", "/web/check", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Form parsing error: ")
	})

	t.Run("Checking Failed", func(t *testing.T) {
		mockChecker := new(MockChecker)
		handlerFunc := ProxyCheckWeb(tmpl, mockChecker)
		formData := "proxies=[192.168.0.1:8080,192.168.0.1:8010]"
		req, err := http.NewRequest("POST", "/web/check", strings.NewReader(formData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		mockChecker.On("AwaitCheck", mock.Anything, mock.Anything).Return([]string{}, errors.New("some error Web"))

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Checking failed: some error Web\n", rr.Body.String())
		mockChecker.AssertExpectations(t)
	})

	t.Run("Valid Proxies", func(t *testing.T) {
		mockChecker := new(MockChecker)
		handlerFunc := ProxyCheckWeb(tmpl, mockChecker)
		formData := "proxies=[192.168.0.1:8080,192.168.0.1:8010]"
		req, err := http.NewRequest("POST", "/web/check", strings.NewReader(formData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		mockChecker.On("AwaitCheck", mock.Anything, mock.Anything).Return([]string{"192.168.0.1:8080", "192.168.0.1:8010"}, nil)

		rr := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "[192.168.0.1:8080 192.168.0.1:8010]", rr.Body.String())
		mockChecker.AssertExpectations(t)
	})
}

func TestProxyCheckForm(t *testing.T) {
	tmpl := template.Must(template.New("proxy_check_form.html.tmpl").Parse("Form"))

	handlerFunc := ProxyCheckForm(tmpl)

	req, err := http.NewRequest("GET", "/form", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Form", rr.Body.String())
}

func TestSendProxiesToChannel(t *testing.T) {
	proxies := []string{"192.168.0.1:8080", "192.168.0.2:8080"}
	proxiesCh := sendProxiesToChannel(proxies)

	for _, expected := range proxies {
		actual := <-proxiesCh
		assert.Equal(t, expected, actual)
	}
	_, ok := <-proxiesCh
	assert.False(t, ok, "channel should be closed")
}
