package handler

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"
	"net/http"
	"time"
)

func WithTimeout(h http.Handler, timeout time.Duration) (http.Handler, error) {
	msg, err := json.Marshal(res.ErrorResponse{
		Code:  apperrors.CodeTimeout,
		Error: "operation has timed out",
	})
	if err != nil {
		return nil, err
	}

	return newContentTypeHandler(http.TimeoutHandler(h, timeout, string(msg))), nil
}

func newContentTypeHandler(h http.Handler) http.Handler {
	return &contentTypeHandler{
		h: h,
	}
}

type contentTypeHandler struct {
	h http.Handler
}

func (h *contentTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(res.HeaderContentTypeKey, res.HeaderContentTypeValue)
	h.h.ServeHTTP(w, r)
}
