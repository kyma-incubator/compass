package model

import (
	"fmt"
	"net/url"
)

// EventingConfiguration missing godoc
type EventingConfiguration struct {
	DefaultURL url.URL
}

// RuntimeEventingConfiguration missing godoc
type RuntimeEventingConfiguration struct {
	EventingConfiguration
}

// NewRuntimeEventingConfiguration missing godoc
func NewRuntimeEventingConfiguration(rawEventURL string) (*RuntimeEventingConfiguration, error) {
	validURL, err := url.Parse(rawEventURL)
	if err != nil {
		return nil, err
	}

	return &RuntimeEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: *validURL,
		},
	}, nil
}

// AppPathURL missing godoc
const AppPathURL = "/%s/v1/events"

// ApplicationEventingConfiguration missing godoc
type ApplicationEventingConfiguration struct {
	EventingConfiguration
}

// NewApplicationEventingConfiguration missing godoc
func NewApplicationEventingConfiguration(runtimeEventURL url.URL, appName string) (*ApplicationEventingConfiguration, error) {
	appEventURL := runtimeEventURL
	if appEventURL.Host != "" {
		appEventURL.Path = fmt.Sprintf(AppPathURL, appName)
	}

	return &ApplicationEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: appEventURL,
		},
	}, nil
}

// NewEmptyApplicationEventingConfig missing godoc
func NewEmptyApplicationEventingConfig() (*ApplicationEventingConfiguration, error) {
	return &ApplicationEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: url.URL{},
		},
	}, nil
}
