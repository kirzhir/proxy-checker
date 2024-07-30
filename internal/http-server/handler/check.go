package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

func ProxyCheck(checker *proxy.Checker) http.HandlerFunc {
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

	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}
