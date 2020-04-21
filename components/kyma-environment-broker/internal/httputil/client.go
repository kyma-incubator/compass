package httputil

import (
	"net/http"
	"time"
)

func NewClient(timeoutSec time.Duration, skipCertVeryfication bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = skipCertVeryfication

	return &http.Client{
		Transport: transport,
		Timeout:   timeoutSec * time.Second,
	}
}
