package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// UnauthorizedMiddleware
func UnauthorizedMiddleware(unauthorizedString string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			recorder := httptest.NewRecorder()

			next.ServeHTTP(recorder, r)
			ctx := r.Context()

			brokerResponseBody, err := ioutil.ReadAll(recorder.Body)
			if err != nil {
				log.C(ctx).Errorf("error while reading recorder body: %s", err.Error())
				writeErrorResponse(ctx, rw, http.StatusInternalServerError, errors.New("internal server error"))
				return
			}

			res := gjson.GetBytes(brokerResponseBody, "description")
			if strings.Contains(res.Str, unauthorizedString) {
				log.C(ctx).Info("unauthorized request.. returning 401")

				writeErrorResponse(ctx, rw, http.StatusUnauthorized, errors.New("unauthorized: insufficient scopes"))
				return
			}

			for key, values := range recorder.Header() {
				for _, v := range values {
					rw.Header().Add(key, v)
				}
			}

			rw.WriteHeader(recorder.Code)
			if _, err := rw.Write(brokerResponseBody); err != nil {
				log.C(ctx).Errorf("error while writing response body: %s", err.Error())
				writeErrorResponse(ctx, rw, http.StatusInternalServerError, errors.New("internal server error"))
			}
		})
	}
}

func writeErrorResponse(ctx context.Context, w http.ResponseWriter, statusResponse int, err error) {
	w.Header().Set("Content-type", "application/json")

	w.WriteHeader(statusResponse)
	errorResp := struct {
		Description string `json:"description"`
	}{Description: err.Error()}
	e := json.NewEncoder(w).Encode(errorResp)
	if e != nil {
		log.C(ctx).Errorf("encoding response error: %s, status: %d, response error: %s", e, statusResponse, errorResp)
	}
}
