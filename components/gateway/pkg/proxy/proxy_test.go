package proxy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestURL(t *testing.T) {
	testCases := []struct {
		requestPath    string
		proxyPath      string
		expectedResult string
	}{
		{
			requestPath:    "/foo/bar/graphql",
			proxyPath:      "/foo",
			expectedResult: "/bar/graphql",
		},
		{
			requestPath:    "/foo/bar/graphql",
			proxyPath:      "/",
			expectedResult: "/foo/bar/graphql",
		},
		{
			requestPath:    "/foo/bar/graphql",
			proxyPath:      "/foo/bar/graphql",
			expectedResult: "/",
		},
	}

	for tN, tC := range testCases {
		t.Run(fmt.Sprintf("%d: Request path %s, when server is proxied on %s", tN, tC.requestPath, tC.proxyPath), func(t *testing.T) {
			result := requestURL(tC.requestPath, tC.proxyPath)
			assert.Equal(t, tC.expectedResult, result)
		})
	}
}
