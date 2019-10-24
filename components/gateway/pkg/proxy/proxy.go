package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Proxy struct {
	targetURL    *url.URL
	reverseProxy *httputil.ReverseProxy
}

func New(targetOrigin, proxyPath string) (*httputil.ReverseProxy, error) {
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
	return &httputil.ReverseProxy{Director: director}, nil
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
