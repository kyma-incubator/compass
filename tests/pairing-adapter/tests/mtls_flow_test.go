package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGettingTokenWithMTLSWorks(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, conf.MTLSPairingAdapterURL, strings.NewReader(`{}`))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respParsed := struct {
		Token string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&respParsed)
	require.NoError(t, err)
	require.NotEmpty(t, respParsed.Token)
}
