package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Validator interface {
	Validate(ctx context.Context) map[string]error
}

func encode(w http.ResponseWriter, r *http.Request, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func renderError(w http.ResponseWriter, r *http.Request, message string, status int) {
	http.Error(w, message, status)
}

func respondWithError(w http.ResponseWriter, r *http.Request, status int, message string, details map[string]error) {
	response := errorResponse{Message: message, Details: details}
	if err := encode(w, r, status, response); err != nil {
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

func (e *errorResponse) MarshalJSON() ([]byte, error) {
	details := make(map[string]string, len(e.Details))
	for key, err := range e.Details {
		details[key] = err.Error()
	}
	return json.Marshal(struct {
		Message string            `json:"message"`
		Details map[string]string `json:"details,omitempty"`
	}{
		Message: e.Message,
		Details: details,
	})
}

func (e *errorResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Message string            `json:"message"`
		Details map[string]string `json:"details,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.Message = raw.Message
	e.Details = make(map[string]error, len(raw.Details))
	for key, problem := range raw.Details {
		e.Details[key] = fmt.Errorf(problem)
	}
	return nil
}
