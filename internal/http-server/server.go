package http_server

import (
	"html/template"
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/http-server/middleware"
	"proxy-checker/internal/proxy"
)

func New(cfg *config.Config, template *template.Template) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		template,
		proxy.NewChecker(cfg.ProxyChecker),
	)

	var serveMux http.Handler = mux
	serveMux = middleware.Logging(serveMux)
	serveMux = middleware.RequestSizing(serveMux)

	return serveMux
}
