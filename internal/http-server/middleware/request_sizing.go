package middleware

import (
	"log/slog"
	"net/http"
)

const MaxUploadSize = 5 * 1024 * 1024 // 5 MB  TODO: Move to a config file

func RequestSizing(next http.Handler) http.Handler {
	log := slog.With(
		slog.String("component", "middleware/request_sizing"),
	)

	log.Info("request_sizing middleware enabled")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > MaxUploadSize {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
		next.ServeHTTP(w, r)
	})
}
