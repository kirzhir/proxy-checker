package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	err := encode(rr, &http.Request{}, http.StatusOK, data)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content type application/json, got %s", contentType)
	}

	expectedBody := `{"key":"value"}`
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, rr.Body.String())
	}
}

func TestDecode(t *testing.T) {
	data := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(data))
	var v map[string]string

	err := decode(req, &v)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	expected := map[string]string{"key": "value"}
	if v["key"] != expected["key"] {
		t.Errorf("expected %v, got %v", expected, v)
	}
}

func TestRenderError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	renderError(rr, req, "Something went wrong", http.StatusBadRequest)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}

	expectedBody := "Something went wrong\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, rr.Body.String())
	}
}

func TestRespondWithError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	details := map[string]error{"field": errors.New("invalid field")}
	respondWithError(rr, req, http.StatusBadRequest, "Validation error", details)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}

	var response errorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Message != "Validation error" {
		t.Errorf("expected message 'Validation error', got %s", response.Message)
	}

	if response.Details["field"].Error() != "invalid field" {
		t.Errorf("expected error 'invalid field' for field, got %s", response.Details["field"].Error())
	}
}

func TestRespondWithSuccess(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	data := map[string]string{"status": "ok"}
	respondWithSuccess(rr, req, data)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", response["status"])
	}
}

func TestErrorResponse_MarshalUnmarshalJSON(t *testing.T) {
	original := errorResponse{
		Message: "Validation error",
		Details: map[string]error{"field": errors.New("invalid field")},
	}

	data, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var decoded errorResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if original.Message != decoded.Message {
		t.Errorf("expected message %s, got %s", original.Message, decoded.Message)
	}

	if original.Details["field"].Error() != decoded.Details["field"].Error() {
		t.Errorf("expected error %s for field, got %s", original.Details["field"].Error(), decoded.Details["field"].Error())
	}
}
