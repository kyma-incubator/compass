package healthz

import (
	"net/http"
)

// NewLivenessHandler returns handler that always return status OK
func NewLivenessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

