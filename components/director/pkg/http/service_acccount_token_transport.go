package http

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// DefaultServiceAccountTokenPath missing godoc
const DefaultServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// InternalAuthorizationHeader missing godoc
const InternalAuthorizationHeader = "X-Authorization"

// NewServiceAccountTokenTransport constructs an serviceAccountTokenTransport
func NewServiceAccountTokenTransport(roundTripper HTTPRoundTripper) *serviceAccountTokenTransport {
	return &serviceAccountTokenTransport{
		roundTripper: roundTripper,
	}
}

// NewServiceAccountTokenTransport constructs an serviceAccountTokenTransport
func NewServiceAccountTokenTransportWithHeader(roundTripper HTTPRoundTripper, headerName string) *serviceAccountTokenTransport {
	return &serviceAccountTokenTransport{
		roundTripper: roundTripper,
		headerName:   headerName,
	}
}

// NewServiceAccountTokenTransportWithPath constructs an serviceAccountTokenTransport with a given path
func NewServiceAccountTokenTransportWithPath(roundTripper HTTPRoundTripper, path string) *serviceAccountTokenTransport {
	return &serviceAccountTokenTransport{
		roundTripper: roundTripper,
		path:         path,
	}
}

// serviceAccountTokenTransport is transport that attaches a kubernetes service account token in the X-Authorization header for internal authentication.
type serviceAccountTokenTransport struct {
	roundTripper HTTPRoundTripper
	path         string
	headerName   string
}

// RoundTrip attaches a kubernetes service account token in the X-Authorization header for internal authentication.
func (tr *serviceAccountTokenTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	path := tr.path
	if len(path) == 0 {
		path = DefaultServiceAccountTokenPath
	}
	token, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read service account token file")
	}

	headerName := InternalAuthorizationHeader
	if tr.headerName != "" {
		headerName = tr.headerName
	}
	r.Header.Set(headerName, "Bearer "+string(token))

	return tr.roundTripper.RoundTrip(r)
}
