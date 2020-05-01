package ias

import (
	"net/http"
)

//go:generate mockery -name=BundleBuilder -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=Bundle -output=automock -outpkg=automock -case=underscore
type (
	BundleBuilder interface {
		NewBundle(identifier string, inputID SPInputID) (Bundle, error)
	}

	Bundle interface {
		FetchServiceProviderData() error
		ServiceProviderName() string
		ServiceProviderType() string
		ServiceProviderExist() bool
		CreateServiceProvider() error
		DeleteServiceProvider() error
		ConfigureServiceProvider() error
		ConfigureServiceProviderType(path string) error
		GenerateSecret() (*ServiceProviderSecret, error)
	}
)

type Builder struct {
	iasClient *Client
	config    Config
}

func NewBundleBuilder(httpClient *http.Client, config Config) BundleBuilder {
	client := NewClient(httpClient, ClientConfig{
		URL:    config.URL,
		ID:     config.UserID,
		Secret: config.UserSecret,
	})

	return &Builder{
		iasClient: client,
		config:    config,
	}
}

func (b *Builder) NewBundle(identifier string, inputID SPInputID) (Bundle, error) {
	if err := inputID.isValid(); err != nil {
		return nil, err
	}
	return NewServiceProviderBundle(identifier, ServiceProviderInputs[inputID], b.iasClient, b.config), nil
}
