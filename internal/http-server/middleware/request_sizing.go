package middleware

import (
	"net/http"
)

const MaxUploadSize = 5 * 1024 * 1024 // 5 MB  TODO: Move to a config file

func RequestSizing() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > MaxUploadSize {
				http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
