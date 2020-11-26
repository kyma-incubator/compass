package healthz

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"

	"github.com/pkg/errors"
)

//go:generate mockery -name=Pinger -output=automock -outpkg=automock -case=underscore
type Pinger interface {
	PingContext(ctx context.Context) error
}

// NewLivenessHandler returns handler that pings DB
func NewLivenessHandler(p Pinger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger := log.C(request.Context())
		err := p.PingContext(request.Context())
		if err != nil {
			logger.Errorf("Got error on checking connection with DB: [%v]", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write([]byte("ok"))
		if err != nil {
			logger.Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}

// NewReadinessHandler returns handler that always return status OK
func NewReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}
