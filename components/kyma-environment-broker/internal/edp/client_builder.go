package edp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	namespaceToken = "%s/oauth2/token"
)

type ClientConfig struct {
	AuthURL   string
	AdminURL  string
	Namespace string
	Secret    string
}

func CreateEDPAdminClient(config ClientConfig, log logrus.FieldLogger) *Client {
	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "edp-namespace.read edp-namespace.update")

	cfg := clientcredentials.Config{
		ClientID:       fmt.Sprintf("edp-namespace;%s", config.Namespace),
		ClientSecret:   config.Secret,
		TokenURL:       fmt.Sprintf(namespaceToken, config.AuthURL),
		EndpointParams: data,
	}
	httpClientOAuth := cfg.Client(context.Background())

	return NewClient(config, httpClientOAuth, log)
}
