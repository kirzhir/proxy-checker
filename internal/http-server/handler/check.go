package handler

import (
	"net/http"
	"proxy-checker/internal/proxy"
)

type errorResponse struct {
	Message string `json:"message"`
}

func ProxyCheck(checker *proxy.Checker) http.HandlerFunc {

	type request []string

	type response []string

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := decode[request](r)

		resp := response(req)

		err = checker.Check(ctx, req[0])
		if err != nil {
			err = encode[errorResponse](w, r, http.StatusOK, errorResponse{Message: err.Error()})
		} else {
			err = encode[response](w, r, http.StatusOK, resp)
		}
	}
}
