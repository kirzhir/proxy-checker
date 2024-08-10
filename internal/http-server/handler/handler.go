package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Validator interface {
	Validate(ctx context.Context) map[string]error
}

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func renderError(w http.ResponseWriter, r *http.Request, message string, status int) {
	http.Error(w, message, status)
}

func respondWithError(w http.ResponseWriter, r *http.Request, status int, message string, details map[string]error) {
	response := errorResponse{Message: message, Details: details}
	if err := encode[errorResponse](w, r, status, response); err != nil {
		slog.Error("Failed to encode error response: " + err.Error())
	}
}

func respondWithSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	if err := encode(w, r, http.StatusOK, data); err != nil {
		slog.Error("Failed to encode success response: " + err.Error())
	}
}

type errorResponse struct {
	Message string           `json:"message"`
	Details map[string]error `json:"details,omitempty"`
}

func (e errorResponse) MarshalJSON() ([]byte, error) {
	details := make(map[string]string, len(e.Details))
	for key, err := range e.Details {
		details[key] = err.Error()
	}
	return json.Marshal(struct {
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	}{
		Message: e.Message,
		Details: details,
	})
}

func (e *errorResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Message  string            `json:"message"`
		Problems map[string]string `json:"details"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.Message = raw.Message
	e.Details = make(map[string]error, len(raw.Problems))
	for key, problem := range raw.Problems {
		e.Details[key] = errors.New(problem)
	}
	return nil
}
