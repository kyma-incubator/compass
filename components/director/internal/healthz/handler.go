package healthz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=Pinger -output=automock -outpkg=automock -case=underscore
type Pinger interface {
	PingContext(ctx context.Context) error
}

// NewHTTPHandler returns function which handles healtz calls
func NewHTTPHandler(p Pinger, log *logrus.Logger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := p.PingContext(request.Context())
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write([]byte(fmt.Sprintf("error: %v", err)))
			if err != nil {
				log.Errorf(errors.Wrapf(err, "while writing to response body on ping failure").Error())
			}
			return
		}
		writer.WriteHeader(200)
		_, err = writer.Write([]byte("ok"))
		if err != nil {
			log.Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}
