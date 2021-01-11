package server_test

import (
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"unsafe"
)

func TestNewAddsAdditionalRoutes(t *testing.T) {
	config := server.DefaultConfig()

	server := server.New(config, []mux.MiddlewareFunc{}, func(router *mux.Router) {
		router.HandleFunc(config.RootAPI+"/test", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})
	})

	AssertRouteExists(t, server, config.RootAPI+"/test")
}

func TestNewAddsSystemLivenessRoutes(t *testing.T) {
	config := server.DefaultConfig()

	var tests = []struct {
		Msg   string
		Route string
	}{
		{
			Msg:   "Health route should exist",
			Route: "/healthz",
		},
		{
			Msg:   "Ready route should exist",
			Route: "/readyz",
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			server := server.New(config, []mux.MiddlewareFunc{})
			AssertRouteExists(t, server, test.Route)
		})
	}
}

func TestNewAddsSystemRoutes(t *testing.T) {
	config := server.DefaultConfig()

	var tests = []struct {
		Msg   string
		Route string
	}{
		{
			Msg:   "Metrics route should exist",
			Route: "/metrics",
		},
		{
			Msg:   "Pprof root route should exist",
			Route: "/debug/pprof/",
		},
		{
			Msg:   "Pprof cmdline route should exist",
			Route: "/debug/pprof/cmdline",
		},
		{
			Msg:   "Pprof profile route should exist",
			Route: "/debug/pprof/profile",
		},
		{
			Msg:   "Pprof symbol route should exist",
			Route: "/debug/pprof/symbol",
		},
		{
			Msg:   "Pprof trace route should exist",
			Route: "/debug/pprof/trace",
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			server := server.New(config, []mux.MiddlewareFunc{})
			AssertRouteExists(t, server, config.RootAPI+test.Route)
		})
	}
}

func AssertRouteExists(t *testing.T, server *server.Server, path string) {
	handler := server.Handler

	router, ok := extractMuxRouter(handler)
	require.True(t, ok)

	match := &mux.RouteMatch{}
	require.True(t, router.Match(&http.Request{
		URL: &url.URL{
			Path: path,
		},
	}, match), match.MatchErr)
}

func extractMuxRouter(handler http.Handler) (*mux.Router, bool) {
	innerHandlerValue := reflect.ValueOf(handler).Elem().FieldByName("h").Elem()
	innerHandlerValue = innerHandlerValue.Elem().FieldByName("h").Elem()
	innerHandlerValue = innerHandlerValue.Elem().FieldByName("handler").Elem()
	innerHandlerValue = innerHandlerValue.Elem().FieldByName("h")
	routerValue := reflect.NewAt(innerHandlerValue.Type(), unsafe.Pointer(innerHandlerValue.UnsafeAddr())).Elem()

	router, ok := routerValue.Interface().(*mux.Router)
	return router, ok
}
