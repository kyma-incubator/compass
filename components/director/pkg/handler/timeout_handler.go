package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
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

	preTimoutLoggingHandler := newTimeoutLoggingHandler(h, timeout, msg)
	timeoutHandler := http.TimeoutHandler(preTimoutLoggingHandler, timeout, string(msg))
	postTimeoutLoggingHandler := newTimeoutLoggingHandler(timeoutHandler, timeout, msg)

	return newContentTypeHandler(postTimeoutLoggingHandler), nil
}

func newTimeoutLoggingHandler(h http.Handler, timeout time.Duration, msg []byte) http.Handler {
	return &timeoutLoggingHandler{
		h:       h,
		timeout: timeout,
		msg:     msg,
	}
}

type timeoutLoggingHandler struct {
	h       http.Handler
	timeout time.Duration
	msg     []byte
}

func (h *timeoutLoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := log.ContextWithLogger(r.Context(), log.LoggerWithCorrelationId(r))

	timoutRW := &timoutLoggingResponseWriter{
		ResponseWriter: w,
		method:         template.HTMLEscapeString(r.Method),
		url:            template.HTMLEscapeString(r.URL.String()),
		timeout:        h.timeout,
		msg:            h.msg,
		requestStart:   time.Now(),
		ctx:            ctx,
	}

	h.h.ServeHTTP(timoutRW, r)
}

type timoutLoggingResponseWriter struct {
	http.ResponseWriter
	method       string
	url          string
	timeout      time.Duration
	msg          []byte
	requestStart time.Time
	ctx          context.Context
}

func (lrw *timoutLoggingResponseWriter) Write(b []byte) (int, error) {
	if bytes.Equal(lrw.msg, b) && time.Since(lrw.requestStart) > lrw.timeout {
		log.C(lrw.ctx).Warnf("%s request to %s timed out after %s", lrw.method, lrw.url, lrw.timeout)
	}

	i, err := lrw.ResponseWriter.Write(b)

	if err != nil && strings.Contains(err.Error(), http.ErrHandlerTimeout.Error()) && time.Since(lrw.requestStart) > lrw.timeout {
		log.C(lrw.ctx).Warnf("Finished processing %s request to %s due to exceeded timeout of %s. Request processing terminated %s after timeout.",
			lrw.method, lrw.url, lrw.timeout, time.Since(lrw.requestStart)-lrw.timeout)
	}
	return i, err
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
