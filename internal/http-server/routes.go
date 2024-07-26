package http_server

import (
	"net/http"
	"proxy-checker/internal/http-server/handler"
	"proxy-checker/internal/proxy"
)

func addRoutes(
	mux *http.ServeMux,
	checker *proxy.Checker,
) {
	//mux.HandleFunc("/healthz", handleHealthzPlease(logger))

	mux.Handle("/api/v1/", ApiMux(checker))
	mux.Handle("/", http.NotFoundHandler())
}

func ApiMux(checker *proxy.Checker) http.Handler {
	apiMux := http.NewServeMux()

	apiMux.Handle("/check", handler.ProxyCheck(checker))

	return http.StripPrefix("/api/v1", apiMux)
}
