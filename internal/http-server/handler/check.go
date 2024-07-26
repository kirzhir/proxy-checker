package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"proxy-checker/internal/config"
	"proxy-checker/internal/proxy"
	"sync"
)

const oneTimeCheckLimit = 100

type response []string

type request []string

func (r request) Valid(ctx context.Context) map[string]error {
	var errors = map[string]error{}

	if len(r) < 1 {
		errors["data"] = fmt.Errorf("request cannot be empty")
	} else if len(r) > oneTimeCheckLimit {
		errors["data"] = fmt.Errorf("request cannot contain more than %d lines", oneTimeCheckLimit)
	}

	return errors
}

func ProxyCheck(checker *proxy.Checker, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := decode[request](r)

		if err != nil {
			responseFail(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if problems := req.Valid(ctx); len(problems) > 0 {
			fmt.Println(problems)
			responseFail(w, r, http.StatusBadRequest, "invalid request", problems)
			return
		}

		proxiesCh := make(chan string, oneTimeCheckLimit)
		for _, p := range req {
			proxiesCh <- p
		}
		close(proxiesCh)

		resp := response{}
		for p := range runChecking(ctx, proxiesCh, cfg.ProxyChecker) {
			resp = append(resp, p)
		}

		responseSuccess(w, r, resp)
	}
}

func runChecking(ctx context.Context, proxiesCh <-chan string, cfg config.ProxyChecker) <-chan string {
	res := make(chan string, len(proxiesCh))

	wg := &sync.WaitGroup{}
	for i := 0; i < oneTimeCheckLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checker := proxy.NewChecker(cfg)

			for {
				select {
				case ch, ok := <-proxiesCh:
					if !ok {
						return
					}

					if err := checker.Check(ctx, ch); err != nil {
						slog.Debug(err.Error())
					} else {
						res <- ch
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	wg.Wait()
	close(res)

	return res
}
