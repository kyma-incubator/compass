package healthz

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=Pinger -output=automock -outpkg=automock -case=underscore
type Pinger interface {
	PingContext(ctx context.Context) error
}

// NewLivenessHandler returns handler that pings DB
func NewLivenessHandler(p Pinger, log *logrus.Logger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := p.PingContext(request.Context())
		if err != nil {
			log.Errorf("Got error on checking connection with DB: [%v]", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write([]byte("ok"))
		if err != nil {
			log.Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}

// NewReadinessHandler returns handler that always return status OK
func NewReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}
