package edp

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	namespaceToken = "%s/oauth2/token"
)

type Config struct {
	AuthURL     string
	AdminURL    string
	Namespace   string
	Secret      string
	Environment string `envconfig:"default=prod"`
	Required    bool   `envconfig:"default=false"`
	Disabled    bool
}

func CreateEDPAdminClient(config Config, log logrus.FieldLogger) *Client {
	cfg := clientcredentials.Config{
		ClientID:     fmt.Sprintf("edp-namespace;%s", config.Namespace),
		ClientSecret: config.Secret,
		TokenURL:     fmt.Sprintf(namespaceToken, config.AuthURL),
		Scopes:       []string{"edp-namespace.read edp-namespace.update"},
	}
	httpClientOAuth := cfg.Client(context.Background())
	httpClientOAuth.Timeout = 30 * time.Second

	return NewClient(config, httpClientOAuth, log)
}
