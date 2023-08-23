package pairing

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// ScenarioGroup represents a token scenario group
type ScenarioGroup struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// RequestData missing godoc
type RequestData struct {
	Application    graphql.Application
	Tenant         string
	ClientUser     string
	ScenarioGroups []ScenarioGroup
}

// ResponseData missing godoc
type ResponseData struct {
	Token string
}
