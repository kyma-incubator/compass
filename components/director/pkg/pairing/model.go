package pairing

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// RequestData missing godoc
type RequestData struct {
	Application    graphql.Application
	Tenant         string
	ClientUser     string
	ScenarioGroups []model.ScenarioGroup
}

// ResponseData missing godoc
type ResponseData struct {
	Token string
}
