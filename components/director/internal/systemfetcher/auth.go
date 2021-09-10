package systemfetcher

import "net/http"

// HeaderTransport missing godoc
type HeaderTransport struct {
	tenantHeaderName string
	base             http.RoundTripper
	tenant           string
}

// RoundTrip missing godoc
func (ht *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(ht.tenantHeaderName, ht.tenant)

	return ht.base.RoundTrip(req)
}
