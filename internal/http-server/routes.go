package http_server

import (
	"html/template"
	"net/http"
	"proxy-checker/internal/http-server/handler"
	"proxy-checker/internal/http-server/middleware"
	"proxy-checker/internal/proxy"
)

func addRoutes(
	mux *http.ServeMux,
	temp *template.Template,
	checker *proxy.Checker,
) {
	mux.Handle("/api/v1/check", middleware.RateLimiting(handler.ProxyCheckApi(checker)))
	mux.Handle("/healthz", handleHealthz())
	mux.Handle("/check", handler.ProxyCheckWeb(temp, checker))
	mux.Handle("/", handler.ProxyCheckForm(temp))
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
