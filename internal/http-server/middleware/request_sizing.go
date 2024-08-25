package middleware

import (
	"log/slog"
	"net/http"
)

func RequestSizing(maxUploadSize int64, next http.Handler) http.Handler {

	slog.Info("request_sizing middleware enabled",
		slog.String("component", "middleware/request_sizing"),
		slog.Int64("size", maxUploadSize),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxUploadSize {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		next.ServeHTTP(w, r)
	})
}
