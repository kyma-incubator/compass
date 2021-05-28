package graphql

type IntegrationSystem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Extended types used by external API

type IntegrationSystemPageExt struct {
	IntegrationSystemPage
	Data []*IntegrationSystemExt `json:"data"`
}

type IntegrationSystemExt struct {
	IntegrationSystem
	Auths []*IntSysSystemAuth `json:"auths"`
}
