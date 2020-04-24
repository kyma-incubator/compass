package httputil

import (
	"net/http"
	"time"
)

func NewClient(timeoutSec time.Duration, skipCertVerification bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = skipCertVerification

	return &http.Client{
		Transport: transport,
		Timeout:   timeoutSec * time.Second,
	}
}
