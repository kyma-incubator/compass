package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	get, err := http.DefaultClient.Get(conf.InfoUrl)
	require.NoError(t, err)

	info := struct {
		Subject string `json:"certSubject"`
		Issuer  string `json:"certIssuer"`
	}{}

	err = json.NewDecoder(get.Body).Decode(&info)
	require.NoError(t, err)

	require.Equal(t, conf.CertSubject, info.Subject)
	require.Equal(t, conf.CertIssuer, info.Issuer)
}

func TestCallingInfoEndpointFailForMethodsOtherThanGet(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, conf.InfoUrl, strings.NewReader("{}"))
	require.NoError(t, err)
	get, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, get.StatusCode)
}
