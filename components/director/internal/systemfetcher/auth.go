package systemfetcher

import "net/http"

type HeaderTransport struct {
	tenantHeaderName string
	base             http.RoundTripper
	tenant           string
}

func (ht *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(ht.tenantHeaderName, ht.tenant)

	return ht.base.RoundTrip(req)
}
