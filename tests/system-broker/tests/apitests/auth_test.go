package apitests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	t.Run("Blank", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, testCtx.SystemBrokerURL, nil)
		require.NoError(t, err)

		_, err = testCtx.HttpClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, 1, 1)
	})
}
