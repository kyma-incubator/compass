package auditlog

import (
	"net/http"
)

type BasicHttpClient struct {
	cl  *http.Client
	cfg BasicAuthConfig
}

func NewBasicAuthClient(cfg BasicAuthConfig, client *http.Client) *BasicHttpClient {
	return &BasicHttpClient{
		cl:  client,
		cfg: cfg,
	}
}

func (cl *BasicHttpClient) Do(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(cl.cfg.User, cl.cfg.Password)
	return cl.cl.Do(req)
}
