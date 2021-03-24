package idtokenprovider

import (
	"crypto/tls"
	"net/http"
)

// TODO temporary fix
var token string

func GetDexToken() (string, error) {
	if token != "" {
		return token, nil
	}

	config, err := NewConfigFromEnv()
	if err != nil {
		return "", err
	}

	token, err = authenticate(config)
	return token, err
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
