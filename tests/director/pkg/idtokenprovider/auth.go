package idtokenprovider

import (
	"crypto/tls"
	"net/http"
)

func GetDexToken() (string, error) {
	config, err := NewConfigFromEnv()
	if err != nil {
		return "", err
	}

	return authenticate(config)
}

func authenticate(config Config) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: config.ClientConfig.TimeoutSeconds,
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	return idTokenProvider.fetchIdToken()
}
