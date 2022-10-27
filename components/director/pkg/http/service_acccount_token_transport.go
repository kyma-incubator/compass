package http

import (
	"net/http"
	"os"

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

// NewServiceAccountTokenTransportWithHeader constructs an serviceAccountTokenTransport with configurable header name
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
	token, err := os.ReadFile(path)
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

// Clone clones the underlying transport
func (tr *serviceAccountTokenTransport) Clone() HTTPRoundTripper {
	return &serviceAccountTokenTransport{
		roundTripper: tr.roundTripper.Clone(),
		path:         tr.path,
		headerName:   tr.headerName,
	}
}

func (tr *serviceAccountTokenTransport) GetTransport() *http.Transport {
	return tr.roundTripper.GetTransport()
}
