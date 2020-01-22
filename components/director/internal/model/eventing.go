package model

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

type EventingConfiguration struct {
	DefaultURL url.URL
}

type RuntimeEventingConfiguration struct {
	EventingConfiguration
}

func NewRuntimeEventingConfiguration(eventURL string) (*RuntimeEventingConfiguration, error) {
	validUrl, err := validateURL(eventURL)
	if err != nil {
		return nil, err
	}

	return &RuntimeEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: validUrl,
		},
	}, nil
}

const ApplicationEventingURLScheme = "https://%s/%s/v1/events"

type ApplicationEventingConfiguration struct {
	EventingConfiguration
}

func NewApplicationEventingConfiguration(runtimeEventURL url.URL, appName string) (*ApplicationEventingConfiguration, error) {
	var appEventURL string

	if runtimeEventURL.Host == "" {
		appEventURL = ""
	} else {
		appEventURL = fmt.Sprintf(ApplicationEventingURLScheme, runtimeEventURL.Host, appName)
	}

	validUrl, err := validateURL(appEventURL)
	if err != nil {
		return nil, errors.Wrap(err, "while validating created application event url")
	}

	return &ApplicationEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: validUrl,
		},
	}, nil
}

func NewEmptyApplicationEventingConfig() (*ApplicationEventingConfiguration, error) {
	emptyURL, err := validateURL("")
	if err != nil {
		return nil, err
	}

	return &ApplicationEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: emptyURL,
		},
	}, nil
}

func validateURL(rawURL string) (url.URL, error) {
	validUrl, err := url.Parse(rawURL)
	if err != nil {
		return url.URL{}, err
	}
	if validUrl == nil {
		return url.URL{}, errors.New("nil url while parsing raw url")
	}
	return *validUrl, nil
}
