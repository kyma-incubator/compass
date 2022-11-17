package http_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/http/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceAccountTokenTransport_RoundTrip(t *testing.T) {
	const tokenValue = "token"

	// GIVEN
	rt := &automock.HTTPRoundTripper{}
	rt.On("RoundTrip", mock.Anything).Return(nil, nil).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)

		token := req.Header.Get(httputil.InternalAuthorizationHeader)
		assert.Equal(t, "Bearer "+tokenValue, token)
	})

	t.Run("sets service account token", func(t *testing.T) {
		tokenFileName := "token"
		err := os.WriteFile(tokenFileName, []byte(tokenValue), os.ModePerm)
		require.NoError(t, err)

		defer func() {
			err := os.Remove(tokenFileName)
			require.NoError(t, err)
		}()

		saTransport := httputil.NewServiceAccountTokenTransportWithPath(rt, tokenFileName)

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		require.NoError(t, err)

		_, err = saTransport.RoundTrip(req)
		assert.NoError(t, err)
	})
}
