package methodnotallowed

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
)

type methodNotAllowedHandler struct {
}

// CreateMethodNotAllowedHandler creates handler that responds when the endpoint is called with method that is not allowed
func CreateMethodNotAllowedHandler() methodNotAllowedHandler {
	return methodNotAllowedHandler{}
}

// ServeHTTP handles the request
func (h methodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httputil.RespondWithError(r.Context(), w, http.StatusMethodNotAllowed, httputil.Error{
		Code:    http.StatusMethodNotAllowed,
		Message: fmt.Sprintf("method %s is not allowed", r.Method),
	})
}
