package readiness

import (
	"net/http"

	"k8s.io/client-go/discovery"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewHTTPHandler(log *logrus.Logger, apiServerClient discovery.ServerVersionInterface) func(writer http.ResponseWriter, request *http.Request) {
	writeResponseFunc := handleResponseFunc(log)
	return func(writer http.ResponseWriter, request *http.Request) {
		_, err := apiServerClient.ServerVersion()
		if err != nil {
			logrus.Errorf("Failed to access API Server: %s.", err.Error())
			writeResponseFunc(writer, http.StatusServiceUnavailable, "Service Unavailable")
			return
		}
		logrus.Debug("Readiness probe passed.")
		writeResponseFunc(writer, http.StatusOK, "ok")
	}
}

func handleResponseFunc(log *logrus.Logger) func(http.ResponseWriter, int, string) {
	return func(writer http.ResponseWriter, statusCode int, body string) {
		writer.WriteHeader(statusCode)
		_, err := writer.Write([]byte(body))
		if err != nil {
			log.Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}
