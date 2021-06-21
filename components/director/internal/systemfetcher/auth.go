package systemfetcher

import "net/http"

type HeaderTransport struct {
	base   http.RoundTripper
	tenant string
}

func (ht *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-zid", ht.tenant)

	return ht.base.RoundTrip(req)
}
