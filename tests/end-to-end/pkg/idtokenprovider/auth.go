package idtokenprovider

import (
	"crypto/tls"
	"net/http"
)

func Authenticate(config Config) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: config.ClientConfig.TimeoutSeconds,
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()
	return token, err
}
