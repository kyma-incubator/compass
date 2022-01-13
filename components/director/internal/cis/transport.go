package cis

import (
	"net/http"
)

type CustomTransport struct {
	http.RoundTripper
	// ... private fields
}

func NewCustomTransport(upstream *http.Transport) *CustomTransport {
	return &CustomTransport{upstream}
}

func (ct *CustomTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for i := 1; i <= 2; i++ {
		resp, err = ct.RoundTripper.RoundTrip(req)
		if resp.StatusCode == http.StatusUnauthorized {
			continue
		} else {
			break
		}
	}

	return resp, err
}
