package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

type Proxy struct {
	targetURL    *url.URL
	reverseProxy *httputil.ReverseProxy
}

func New(targetOrigin, proxyPath string, transport http.RoundTripper) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(targetOrigin)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing URL %s", targetOrigin)
	}

	targetQuery := targetURL.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = requestURL(req.URL.Path, proxyPath)
		req.Host = targetURL.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			log.C(req.Context()).WithError(err).Errorf("Error while proxying request to %q", req.URL.String())
			rw.WriteHeader(http.StatusBadGateway)
		},
	}, nil
}

func requestURL(requestPath, proxyPath string) string {
	if proxyPath == "/" {
		return requestPath
	}

	trimmedPath := strings.TrimPrefix(requestPath, proxyPath)
	if trimmedPath == "" {
		return "/"
	}

	return trimmedPath
}
