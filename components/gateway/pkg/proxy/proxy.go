package proxy

import (
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func New(targetOrigin string) (*httputil.ReverseProxy, error) {
	targetUrl, err := url.Parse(targetOrigin)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing URL %s", targetOrigin)
	}

	targetQuery := targetUrl.RawQuery
	p := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetUrl.Scheme
			req.URL.Host = targetUrl.Host
			req.URL.Path = singleJoiningSlash(targetUrl.Path, req.URL.Path)
			req.Host = targetUrl.Host
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
	}

	return p, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
