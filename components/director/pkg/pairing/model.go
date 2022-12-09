package pairing

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

// RequestData missing godoc
type RequestData struct {
	Application    graphql.Application
	Tenant         string
	ClientUser     string
	ScenarioGroups []string
}

// ResponseData missing godoc
type ResponseData struct {
	Token string
}
