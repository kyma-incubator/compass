package tests

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	plusDelimiter  = "+"
	commaDelimiter = ","
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

	expectedIssuer, expectedSubject := getExpectedFields(t)

	require.Equal(t, expectedSubject, info.Subject)
	require.Equal(t, expectedIssuer, info.Issuer)
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

func getExpectedFields(t *testing.T) (string, string) {
	clientCert := cc.Get()[conf.ExternalClientCertSecretName]
	require.NotNil(t, clientCert)
	require.NotEmpty(t, clientCert.Certificate)

	parsedClientCert, err := x509.ParseCertificate(clientCert.Certificate[0])
	require.NoError(t, err)

	return replaceDelimiter(parsedClientCert.Issuer.String()), replaceDelimiter(parsedClientCert.Subject.String())
}

func replaceDelimiter(input string) string {
	return strings.ReplaceAll(input, plusDelimiter, commaDelimiter)
}
