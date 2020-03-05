package avs

import (
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"sync"
	"sync/atomic"
)

type clientHolder struct {
	httpClient        *http.Client
	avsConfig         Config
	clientInitialized uint32
	mu                sync.Mutex
}

func newClientHolder(avsConfig Config) *clientHolder {
	return &clientHolder{
		avsConfig: avsConfig,
	}
}

func (ch *clientHolder) getClient(logger logrus.FieldLogger) (*http.Client, error) {
	if atomic.LoadUint32(&ch.clientInitialized) == 1 {
		return ch.httpClient, nil
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	if atomic.LoadUint32(&ch.clientInitialized) == 0 {
		logger.Infof("creating the lazy client for avs..")
		httpClient, err := ch.createClient()
		if err != nil {
			return nil, err
		}
		ch.httpClient = httpClient
		atomic.StoreUint32(&ch.clientInitialized, 1)
	}
	return ch.httpClient, nil

}

func (ch *clientHolder) createClient() (*http.Client, error) {
	config := &oauth2.Config{
		ClientID: ch.avsConfig.OauthClientId,
		Endpoint: oauth2.Endpoint{
			TokenURL:  ch.avsConfig.OauthTokenEndpoint,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
	initialToken, err := config.PasswordCredentialsToken(context.Background(),
		ch.avsConfig.OauthUsername, ch.avsConfig.OauthPassword)

	if err != nil {
		return nil, err
	}
	return config.Client(context.Background(), initialToken), nil
}
