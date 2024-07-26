package http_server

import (
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/http-server/handler"
	"proxy-checker/internal/proxy"
)

func addRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	checker *proxy.Checker,
) {
	mux.Handle("/api/v1/check", handler.ProxyCheck(checker, cfg))
	mux.Handle("/healthz", handleHealthz())
	mux.Handle("/", http.NotFoundHandler())
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
