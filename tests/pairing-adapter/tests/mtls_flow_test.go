package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	director_http "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/stretchr/testify/require"
)

func TestGettingTokenWithMTLSWorks(t *testing.T) {
	//TODO: Change templates to be real
	req, err := http.NewRequest(http.MethodPost, conf.MTLSPairingAdapterURL, strings.NewReader(`{}`))
	require.NoError(t, err)

	client := http.Client{
		Transport: director_http.NewServiceAccountTokenTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respParsed := struct {
		Token string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&respParsed)
	require.NoError(t, err)
	require.NotEmpty(t, respParsed.Token)
}
