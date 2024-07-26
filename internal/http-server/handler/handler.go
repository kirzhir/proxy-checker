package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type Validator interface {
	Valid(ctx context.Context) map[string]error
}

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func render(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("render HTML: %w", err)
	}
	return nil
}

func responseFail(w http.ResponseWriter, r *http.Request, status int, message string, problems map[string]error) {
	if err := encode[errorResponse](w, r, status, errorResponse{Message: message, Problems: problems}); err != nil {
		slog.Error(err.Error())
	}
}

func responseSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	if err := encode(w, r, http.StatusOK, data); err != nil {
		slog.Error(err.Error())
	}
}

type errorResponse struct {
	Message  string
	Problems map[string]error
}

func (e errorResponse) MarshalJSON() ([]byte, error) {
	problems := make(map[string]string, len(e.Problems))
	for key, err := range e.Problems {
		problems[key] = err.Error()
	}

	return json.Marshal(struct {
		Message  string            `json:"message"`
		Problems map[string]string `json:"problems"`
	}{
		Message:  e.Message,
		Problems: problems,
	})
}
