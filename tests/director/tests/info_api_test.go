package tests

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(infoEndpoint())
	require.NoError(t, err)

	info := struct {
		Subject string `json:"certSubject"`
		Issuer  string `json:"certIssuer"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&info)
	require.NoError(t, err)

	require.Equal(t, conf.CertSubject, info.Subject)
	require.Equal(t, conf.CertIssuer, info.Issuer)
}

func TestCallingInfoEndpointFailForMethodsOtherThanGet(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest(http.MethodPost, infoEndpoint(), strings.NewReader("{}"))
	require.NoError(t, err)
	resp, err := client.Do(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func infoEndpoint() string {
	directorPath := "/director"
	infoEndpoint := conf.DirectorUrl
	if strings.Contains(infoEndpoint, directorPath) {
		infoEndpoint = infoEndpoint[:strings.Index(infoEndpoint, directorPath)]
		infoEndpoint = infoEndpoint + conf.InfoUrl
	}
	return infoEndpoint
}
