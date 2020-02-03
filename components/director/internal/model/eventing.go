package model

import (
	"fmt"
	"net/url"
)

type EventingConfiguration struct {
	DefaultURL url.URL
}

type RuntimeEventingConfiguration struct {
	EventingConfiguration
}

func NewRuntimeEventingConfiguration(rawEventURL string) (*RuntimeEventingConfiguration, error) {
	validUrl, err := url.Parse(rawEventURL)
	if err != nil {
		return nil, err
	}

	return &RuntimeEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: *validUrl,
		},
	}, nil
}

const AppPathURL = "/%s/v1/events"

type ApplicationEventingConfiguration struct {
	EventingConfiguration
}

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

func NewEmptyApplicationEventingConfig() (*ApplicationEventingConfiguration, error) {
	return &ApplicationEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: url.URL{},
		},
	}, nil
}
