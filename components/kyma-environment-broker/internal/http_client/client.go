package http_client

import (
	"crypto/tls"
	"net/http"
	"time"
)

func NewHTTPClient(timeoutSec time.Duration, skipCertVeryfication bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVeryfication},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeoutSec * time.Second,
	}
}
