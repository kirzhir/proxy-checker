package http_server

import (
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/http-server/middleware"
	"proxy-checker/internal/proxy"
)

func New(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		proxy.NewChecker(cfg.ProxyChecker),
	)

	var serveMux http.Handler = mux
	serveMux = middleware.Logging(serveMux)
	serveMux = middleware.RequestSizing(serveMux)

	return serveMux
}
