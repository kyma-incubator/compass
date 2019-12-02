package model

type EventingConfiguration struct {
	DefaultURL string
}

type RuntimeEventingConfiguration struct {
	EventingConfiguration
}

func NewRuntimeEventingConfiguration(url string) *RuntimeEventingConfiguration {
	return &RuntimeEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: url,
		},
	}
}

type ApplicationEventingConfiguration struct {
	EventingConfiguration
}

func NewApplicationEventingConfiguration(url string) *ApplicationEventingConfiguration {
	return &ApplicationEventingConfiguration{
		EventingConfiguration: EventingConfiguration{
			DefaultURL: url,
		},
	}
}
