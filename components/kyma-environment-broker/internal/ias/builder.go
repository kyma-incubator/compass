package ias

import (
	"net/http"
)

//go:generate mockery -name=BundleBuilder -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=Bundle -output=automock -outpkg=automock -case=underscore
type (
	BundleBuilder interface {
		NewBundle(identifier string) Bundle
	}

	Bundle interface {
		FetchServiceProviderData() error
		ServiceProviderName() string
		ServiceProviderExist() bool
		CreateServiceProvider() error
		DeleteServiceProvider() error
		ConfigureServiceProvider() error
		ConfigureServiceProviderType(path string) error
		GenerateSecret() (*ServiceProviderSecret, error)
	}
)

type Builder struct {
	httpClient *http.Client
	config     Config
}

func NewBundleBuilder(httpClient *http.Client, config Config) BundleBuilder {
	return &Builder{
		httpClient: httpClient,
		config:     config,
	}
}

func (b *Builder) NewBundle(identifier string) Bundle {
	client := NewClient(b.httpClient, ClientConfig{
		URL:    b.config.URL,
		ID:     b.config.UserID,
		Secret: b.config.UserSecret,
	})

	return NewServiceProviderBundle(identifier, client, b.config)
}
