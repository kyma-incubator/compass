package accessstrategy_test

import (
	"net/http"
)

var expectedResp = &http.Response{
	StatusCode: http.StatusOK,
	Body:       nil,
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}
