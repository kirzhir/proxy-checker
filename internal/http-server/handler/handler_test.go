package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockValidator struct {
	valid bool
}

func (m mockValidator) Valid(ctx context.Context) map[string]error {
	if m.valid {
		return nil
	}
	return map[string]error{"field": errors.New("invalid")}
}

func TestEncode(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	data := map[string]string{"foo": "bar"}

	err := encode(w, r, http.StatusOK, data)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, data, response)
}

func TestDecode(t *testing.T) {
	data := map[string]string{"foo": "bar"}
	body, err := json.Marshal(data)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))

	var result map[string]string
	result, err = decode[map[string]string](r)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestRenderFail(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	msg := "error message"
	status := http.StatusBadRequest

	renderFail(w, r, msg, status)
	assert.Equal(t, status, w.Result().StatusCode)
	assert.Equal(t, msg+"\n", w.Body.String())
}

func TestRender(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	tmpl, err := template.New("test").Parse("<html><body>{{.}}</body></html>")
	require.NoError(t, err)
	data := "Hello, World!"

	err = render(w, r, tmpl, data)
	require.NoError(t, err)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), data)
}

func TestResponseFail(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	status := http.StatusBadRequest
	message := "fail message"
	problems := map[string]error{"field": errors.New("invalid")}

	responseFail(w, r, status, message, problems)
	assert.Equal(t, status, w.Result().StatusCode)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var response errorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, message, response.Message)
	assert.Equal(t, len(problems), len(response.Problems))
	assert.Equal(t, problems["field"].Error(), response.Problems["field"].Error())
}

func TestResponseSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	data := map[string]string{"foo": "bar"}

	responseSuccess(w, r, data)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, data, response)
}
