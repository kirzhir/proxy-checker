package http_server

import (
	"net/http"
	"proxy-checker/internal/config"
)

func addRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
) {
	//mux.HandleFunc("/healthz", handleHealthzPlease(logger))

	mux.Handle("/api/v1/", ApiMux(cfg))
	mux.Handle("/", http.NotFoundHandler())
}

func ApiMux(cfg *config.Config) http.Handler {
	apiMux := http.NewServeMux()

	apiMux.Handle("/check", handleProxyCheck(cfg.ProxyChecker))

	return http.StripPrefix("/api/v1", apiMux)
}
