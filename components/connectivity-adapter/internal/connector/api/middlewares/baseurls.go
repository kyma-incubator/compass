package middlewares

import (
	"net/http"
)

//go:generate mockery -name=RuntimeBaseURLProvider -output=automock -outpkg=automock -case=underscore
type RuntimeBaseURLProvider interface {
	EventServiceBaseURL() (string, error)
}

type baseURLsMiddleware struct {
	connectivityAdapterBaseURL string
	runtimeBaseURLProvider     RuntimeBaseURLProvider
}

func NewBaseURLsMiddleware(connectivityAdapterBaseURL string, runtimeBaseURLProvider RuntimeBaseURLProvider) baseURLsMiddleware {
	return baseURLsMiddleware{
		connectivityAdapterBaseURL: connectivityAdapterBaseURL,
		runtimeBaseURLProvider:     runtimeBaseURLProvider,
	}
}

func (b baseURLsMiddleware) GetBaseUrls(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		eventServiceBaseURL, err := b.runtimeBaseURLProvider.EventServiceBaseURL()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		baseURLs := BaseURLs{
			ConnectivityAdapterBaseURL: b.connectivityAdapterBaseURL,
			EventServiceBaseURL:        eventServiceBaseURL,
		}

		context := PutIntoContext(r.Context(), BaseURLsKey, baseURLs)

		r = r.WithContext(context)
		handler.ServeHTTP(w, r)
	})
}
