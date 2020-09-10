package readiness

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewHTTPHandler(log *logrus.Logger, isCacheLoaded *atomicBool, cacheNotificationCh <-chan struct{}) func(writer http.ResponseWriter, request *http.Request) {
	writeResponseFunc := handleResponseFunc(log)
	go handleNotification(log, cacheNotificationCh, isCacheLoaded)

	return func(writer http.ResponseWriter, request *http.Request) {
		if !isCacheLoaded.getValue() {
			log.Debug("Needed certificates are still not loaded")
			writeResponseFunc(writer, http.StatusServiceUnavailable, "Service Unavailable")
			return
		}

		logrus.Debug("Readiness probe passed. All needed certificates are loaded")
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

func handleNotification(log *logrus.Logger, cacheNotificationCh <-chan struct{}, isCacheLoaded *atomicBool) {
	<-cacheNotificationCh
	log.Info("Received notification, cache is ready")
	isCacheLoaded.setValue(true)
}
