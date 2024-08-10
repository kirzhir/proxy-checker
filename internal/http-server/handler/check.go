package handler

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"proxy-checker/internal/proxy"
	"strings"
)

const maxProxies = 100

var (
	errEmptyRequest   = fmt.Errorf("request cannot be empty")
	errTooManyProxies = fmt.Errorf("request cannot contain more than %d proxies", maxProxies)
)

type ProxyRequest []string

func (req ProxyRequest) Validate(ctx context.Context) map[string]error {

	errors := map[string]error{}
	if len(req) == 0 {
		errors["proxies"] = errEmptyRequest
	} else if len(req) > maxProxies {
		errors["proxies"] = errTooManyProxies
	}

	return errors
}

func ProxyCheckAPI(checker proxy.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var request ProxyRequest
		if err := decode(r, &request); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Failed to decode request: "+err.Error(), nil)
			return
		}

		if errors := request.Validate(ctx); len(errors) > 0 {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request", errors)
			return
		}

		result, err := checker.AwaitCheck(ctx, sendProxiesToChannel(request))
		if err != nil {
			respondWithError(w, r, http.StatusInternalServerError, "Proxy check failed: "+err.Error(), nil)
			return
		}

		respondWithSuccess(w, r, result)
	}
}

func ProxyCheckWeb(template *template.Template, checker proxy.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := r.ParseForm(); err != nil {
			renderError(w, r, fmt.Sprintf("Form parsing error: %s", err.Error()), http.StatusBadRequest)
			return
		}

		proxies := strings.TrimSpace(r.FormValue("proxies"))
		if proxies == "" {
			renderError(w, r, "Missing 'proxies' parameter", http.StatusBadRequest)
			return
		}

		request := ProxyRequest(strings.Split(proxies, "\n"))
		result, err := checker.AwaitCheck(ctx, sendProxiesToChannel(request))
		if err != nil {
			renderError(w, r, fmt.Sprintf("Checking failed: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		if err = template.ExecuteTemplate(w, "proxies_table.html.tmpl", result); err != nil {
			renderError(w, r, "Failed to render result template: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func ProxyCheckForm(template *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := template.ExecuteTemplate(w, "proxy_check_form.html.tmpl", nil); err != nil {
			slog.Error("Failed to render form template: " + err.Error())
		}
	}
}

func sendProxiesToChannel(proxies []string) chan string {
	proxiesCh := make(chan string, len(proxies))
	for _, p := range proxies {
		proxiesCh <- p
	}
	close(proxiesCh)
	return proxiesCh
}
