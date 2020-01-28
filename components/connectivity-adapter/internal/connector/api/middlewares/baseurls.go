package middlewares

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
)

//go:generate mockery -name=RuntimeBaseURLProvider -output=automock -outpkg=automock -case=underscore
type RuntimeBaseURLProvider interface {
	EventServiceBaseURL() (string, error)
}

type baseURLsMiddleware struct {
	connectivityAdapterBaseURL     string
	connectivityAdapterMTLSBaseURL string
	runtimeBaseURLProvider         RuntimeBaseURLProvider
}

func NewBaseURLsMiddleware(connectivityAdapterBaseURL string, connectivityAdapterMTLSBaseURL string, runtimeBaseURLProvider RuntimeBaseURLProvider) baseURLsMiddleware {
	return baseURLsMiddleware{
		connectivityAdapterBaseURL:     connectivityAdapterBaseURL,
		runtimeBaseURLProvider:         runtimeBaseURLProvider,
		connectivityAdapterMTLSBaseURL: connectivityAdapterMTLSBaseURL,
	}
}

func (bm baseURLsMiddleware) GetBaseUrls(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		eventServiceBaseURL, err := bm.runtimeBaseURLProvider.EventServiceBaseURL()
		if err != nil {
			reqerror.WriteError(w, err, apperrors.CodeInternal)

			return
		}

		baseURLs := BaseURLs{
			ConnectivityAdapterBaseURL:     bm.connectivityAdapterBaseURL,
			ConnectivityAdapterMTLSBaseURL: bm.connectivityAdapterMTLSBaseURL,
			EventServiceBaseURL:            eventServiceBaseURL,
		}

		context := PutIntoContext(r.Context(), BaseURLsKey, baseURLs)

		r = r.WithContext(context)
		handler.ServeHTTP(w, r)
	})
}
