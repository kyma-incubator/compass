package graphql

// IntegrationSystem missing godoc
type IntegrationSystem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// IntegrationSystemPageExt is an extended types used by external API
type IntegrationSystemPageExt struct {
	IntegrationSystemPage
	Data []*IntegrationSystemExt `json:"data"`
}

// IntegrationSystemExt missing godoc
type IntegrationSystemExt struct {
	IntegrationSystem
	Auths []*IntSysSystemAuth `json:"auths"`
}
