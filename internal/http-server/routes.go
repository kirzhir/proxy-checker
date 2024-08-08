package http_server

import (
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"html/template"
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/http-server/handler"
	"proxy-checker/internal/http-server/middleware"
	"proxy-checker/internal/proxy"
)

func New(cfg *config.Config, template *template.Template) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Logging())
	mux.Use(chi_middleware.RequestSize(cfg.MaxRequestSize))

	addRoutes(
		mux,
		template,
		proxy.NewChecker(cfg.ProxyChecker),
	)

	return mux
}

func addRoutes(
	mux *chi.Mux,
	temp *template.Template,
	checker proxy.Checker,
) {

	mux.Mount("/debug", chi_middleware.Profiler())

	mux.Route("/api/v1/check", func(r chi.Router) {
		r.Use(middleware.RateLimiting())
		r.Post("/", handler.ProxyCheckAPI(checker))
	})

	mux.Route("/check", func(r chi.Router) {
		r.Use(middleware.RateLimiting())
		r.Post("/", handler.ProxyCheckWeb(temp, checker))
	})

	mux.Get("/healthz", handleHealthz())
	mux.Get("/", handler.ProxyCheckForm(temp))
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("I'm alive"))
	}
}
