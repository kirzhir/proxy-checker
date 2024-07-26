package http_server

import (
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/http-server/handler"
	"proxy-checker/internal/http-server/middleware"
	"proxy-checker/internal/proxy"
)

func New(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		cfg,
	)

	var serveMux http.Handler = mux
	serveMux = requestSizing(serveMux)
	serveMux = logging(serveMux)

	return serveMux
}

func handleProxyCheck(cfg config.ProxyChecker) http.Handler {
	return handler.ProxyCheck(proxy.NewChecker(cfg.API, cfg.Timeout))
}

func logging(handler http.Handler) http.Handler {
	return middleware.Logging()(handler)
}

func requestSizing(handler http.Handler) http.Handler {
	return middleware.RequestSizing()(handler)
}
