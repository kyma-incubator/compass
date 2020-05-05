package azure

import (
	"net/http"
	"net/http/httputil"

	"github.com/Azure/go-autorest/autorest"
	"go.uber.org/zap"
)

func LogRequest(logger *zap.SugaredLogger, body bool) autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err != nil {
				logger.Error(err)
				return nil, err
			}

			if dump, err := httputil.DumpRequestOut(r, body); err == nil {
				logger.Debug(string(dump))
			}

			return r, nil
		})
	}
}

func LogResponse(logger *zap.SugaredLogger, body bool) autorest.RespondDecorator {
	return func(p autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(r *http.Response) error {
			if err := p.Respond(r); err != nil {
				logger.Error(err)
				return err
			}

			if dump, err := httputil.DumpResponse(r, body); err == nil {
				if r.StatusCode != http.StatusOK {
					logger.Error(string(dump))
				} else {
					logger.Debug(string(dump))
				}
			}

			return nil
		})
	}
}
