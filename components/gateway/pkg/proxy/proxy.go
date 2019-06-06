package proxy

import (
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	targetURL    *url.URL
	reverseProxy *httputil.ReverseProxy
}

func New(targetOrigin string) (*Proxy, error) {
	targetURL, err := url.Parse(targetOrigin)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing URL %s", targetOrigin)
	}

	return &Proxy{
		targetURL:    targetURL,
		reverseProxy: httputil.NewSingleHostReverseProxy(targetURL),
	}, nil
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.Host = p.targetURL.Host
	p.reverseProxy.ServeHTTP(rw, req)
}
