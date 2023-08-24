package client

import "github.com/kyma-incubator/compass/components/instance-creator/internal/config"

type client struct {
	cfg            *config.Config
	callerProvider *CallerProvider
}

func NewClient(cfg *config.Config, callerProvider *CallerProvider) *client {
	return &client{
		cfg:            cfg,
		callerProvider: callerProvider,
	}
}
