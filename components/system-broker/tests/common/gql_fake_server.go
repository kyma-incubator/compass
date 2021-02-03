package common

import (
	"net/http"
	"net/http/httptest"
)

type GqlFakeServer struct {
	server *httptest.Server
}

func NewGqlFakeServer(h http.Handler) *GqlFakeServer {
	return &GqlFakeServer{httptest.NewServer(h)}
}

func (g *GqlFakeServer) Close() {
	g.server.Close()
}

func (g *GqlFakeServer) URL() string {
	return g.server.URL
}
