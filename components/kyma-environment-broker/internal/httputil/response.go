package httputil

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Writer provides syntactic sugar for writing http responses.
// Works in two modes:
//   * devMode: true - returns a given error in the response under `details` field
//   * devMode: false - only log the given error in context of the requestID but do not return it in response
type Writer struct {
	log     logrus.FieldLogger
	devMode bool
}

// NewResponseWriter returns new instance of Writer
func NewResponseWriter(log logrus.FieldLogger, devMode bool) *Writer {
	return &Writer{
		log:     log,
		devMode: devMode,
	}
}

// NotFound writes standard NotFound response to given ResponseWriter.
func (w *Writer) NotFound(rw http.ResponseWriter, r *http.Request, err error, context string) {
	w.writeError(rw, r, ErrorDTO{
		Status:  http.StatusNotFound,
		Message: "Whoops! We can't find what you're looking for. Please try again.",
		Details: fmt.Sprintf("%s: %s", context, err),
	})
}

// InternalServerError writes standard InternalServerError response to given ResponseWriter.
func (w *Writer) InternalServerError(rw http.ResponseWriter, r *http.Request, err error, context string) {
	w.writeError(rw, r, ErrorDTO{
		Status:  http.StatusInternalServerError,
		Message: "Something went very wrong. Please try again.",
		Details: fmt.Sprintf("%s: %s", context, err),
	})
}

func (w *Writer) writeError(rw http.ResponseWriter, r *http.Request, errDTO ErrorDTO) {
	errDTO.RequestID = r.Header.Get("X-Request-Id")
	if !w.devMode {
		w.log.WithField("request-id", errDTO.RequestID).Error(errDTO.Details)
		errDTO.Details = ""
	}

	if err := JSONEncodeWithCode(rw, errDTO, errDTO.Status); err != nil {
		w.log.WithField("request-id", errDTO.RequestID).Errorf("while encoding error DTO: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
