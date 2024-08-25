package http_server

import (
	"html/template"
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/http-server/handler"
	"proxy-checker/internal/http-server/middleware"
	"proxy-checker/internal/proxy"
	"time"
)

func New(cfg *config.Config, template *template.Template) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		template,
		proxy.NewChecker(cfg.ProxyChecker),
	)

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestSizing(cfg.MaxRequestSize, h)

	return h
}

func addRoutes(
	mux *http.ServeMux,
	temp *template.Template,
	checker proxy.Checker,
) {
	mux.Handle("POST /api/v1/check", middleware.RateLimiting(3*time.Minute, handler.ProxyCheckAPI(checker)))
	mux.Handle("POST /check", middleware.RateLimiting(3*time.Minute, handler.ProxyCheckWeb(temp, checker)))
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("GET /ip", handleIP())
	mux.Handle("GET /", handler.ProxyCheckForm(temp))
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("I'm alive"))
	}
}

func handleIP() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.RemoteAddr))
	}
}
