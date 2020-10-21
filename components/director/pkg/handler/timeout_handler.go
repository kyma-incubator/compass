package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

const (
	HeaderContentTypeKey   = "Content-Type"
	HeaderContentTypeValue = "application/json;charset=UTF-8"
)

func WithTimeout(h http.Handler, timeout time.Duration) (http.Handler, error) {
	msg, err := json.Marshal(apperrors.NewOperationTimeoutError())
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
	w.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)
	h.h.ServeHTTP(w, r)
}
