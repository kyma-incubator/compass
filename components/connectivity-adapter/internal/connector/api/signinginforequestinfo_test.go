package api

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestHandler_Delete(t *testing.T) {

	t.Run("Should get Signing Request Info", func(t *testing.T) {
		// given

		// when

		// then

	})

	t.Run("Should return error when failed to call Compass Connector", func(t *testing.T) {
		// given

		// when

		// then

	})

	// TODO: Check if it is needed
	// TODO check what is the response from GraphQL client if the header is missing
	t.Run("Should return error when Client-Id-From-Token not passed", func(t *testing.T) {
		// given

		// when

		// then
	})

}

func closeRequestBody(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
