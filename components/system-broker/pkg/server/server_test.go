package server_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewAddsAdditionalRoutes(t *testing.T) {
	config := server.DefaultConfig()
	uuid := uuid.NewService()

	server := server.New(config, uuid, func(router *mux.Router) {
		router.HandleFunc(config.RootAPI+"/test", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})
	})

	AssertRouteExists(t, server, config.RootAPI+"/test")
}

func TestNewAddsSystemRoutes(t *testing.T) {
	config := server.DefaultConfig()
	uuid := uuid.NewService()

	var tests = []struct {
		Msg   string
		Route string
	}{
		{
			Msg:   "Metrics route should exist",
			Route: "/metrics",
		},
		{
			Msg:   "Health route should exist",
			Route: "/healthz",
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
			server := server.New(config, uuid)
			AssertRouteExists(t, server, config.RootAPI+test.Route)
		})
	}
}

func AssertRouteExists(t *testing.T, server *server.Server, path string) {
	router, ok := server.Handler.(*mux.Router)
	require.True(t, ok)

	match := &mux.RouteMatch{}
	require.True(t, router.Match(&http.Request{
		URL: &url.URL{
			Path: path,
		},
	}, match), match.MatchErr)
}
