package handler

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"proxy-checker/internal/proxy"
	"strings"
	"sync"
)

const oneTimeCheckLimit = 100

type response []string

type request []string

func (r request) Valid(_ context.Context) map[string]error {
	var errors = map[string]error{}

	if len(r) < 1 {
		errors["data"] = fmt.Errorf("request cannot be empty")
	} else if len(r) > oneTimeCheckLimit {
		errors["data"] = fmt.Errorf("request cannot contain more than %d lines", oneTimeCheckLimit)
	}

	return errors
}

func ProxyCheckApi(checker *proxy.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := decode[request](r)

		if err != nil {
			responseFail(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if problems := req.Valid(ctx); len(problems) > 0 {
			responseFail(w, r, http.StatusBadRequest, "invalid request", problems)
			return
		}

		proxiesCh := make(chan string, len(req))
		for _, p := range req {
			proxiesCh <- p
		}
		close(proxiesCh)

		resp := response{}
		for p := range runChecking(ctx, proxiesCh, checker) {
			resp = append(resp, p)
		}

		responseSuccess(w, r, resp)
	}
}

func ProxyCheckForm(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := temp.ExecuteTemplate(w, "proxy_check_form.html.tmpl", nil); err != nil {
			slog.Error(err.Error())
		}
	}
}

func ProxyCheckWeb(temp *template.Template, checker *proxy.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := r.ParseForm(); err != nil {
			renderFail(w, r, fmt.Sprintf("Error parsing form: %s", err.Error()), http.StatusBadRequest)
			return
		}

		proxies := r.FormValue("proxies")
		if proxies == "" {
			renderFail(w, r, "Proxies parameter is missing", http.StatusBadRequest)
			return
		}

		req := strings.Split(proxies, "\n")
		proxiesCh := make(chan string, len(req))
		for _, p := range req {
			proxiesCh <- p
		}
		close(proxiesCh)

		resp := response{}
		for p := range runChecking(ctx, proxiesCh, checker) {
			resp = append(resp, p)
		}

		if err := temp.ExecuteTemplate(w, "proxies_table.html.tmpl", resp); err != nil {
			renderFail(w, r, err.Error(), http.StatusInternalServerError)
		}
	}
}

func runChecking(ctx context.Context, proxiesCh <-chan string, checker *proxy.Checker) <-chan string {
	proxiesNum := len(proxiesCh)
	res := make(chan string, proxiesNum)

	wg := &sync.WaitGroup{}
	for i := 0; i < proxiesNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case ch, ok := <-proxiesCh:
					if !ok {
						return
					}

					if p, err := checker.Check(ctx, ch); err != nil {
						slog.Debug(err.Error())
					} else {
						res <- p
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}
