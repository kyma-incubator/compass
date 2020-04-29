package ias

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=BundleBuilder -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=Bundle -output=automock -outpkg=automock -case=underscore
type (
	BundleBuilder interface {
		NewBundle(identifier string, inputID string) (Bundle, error)
	}

	Bundle interface {
		FetchServiceProviderData() error
		ServiceProviderName() string
		ServiceProviderExist() bool
		CreateServiceProvider() error
		DeleteServiceProvider() error
		ConfigureServiceProvider() error
		ConfigureServiceProviderType(path string) error
		GetProvisioningOverrides() (string, []*gqlschema.ConfigEntryInput)
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

func (b *Builder) NewBundle(identifier string, inputID string) (Bundle, error) {
	spParams, exist := ServiceProviderInputs[inputID]
	if !exist {
		return nil, errors.Errorf("Invalid Service Provider input ID: %s", inputID)
	}
	return NewServiceProviderBundle(identifier, spParams, b.iasClient, b.config), nil
}
