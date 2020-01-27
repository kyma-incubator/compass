package service

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/pkg/errors"
)

type serviceManagerProvider struct {
	gqlProvider gqlcli.Provider
}

func NewServiceManagerProvider(gqlProvider gqlcli.Provider) *serviceManagerProvider {
	return &serviceManagerProvider{gqlProvider: gqlProvider}
}

func (s *serviceManagerProvider) ForRequest(r *http.Request) (ServiceManager, error) {
	appDetails, err := appdetails.LoadFromContext(r.Context())
	if err != nil {
		return nil, errors.Wrap(err, "while loading Application details from context")
	}

	// TODO: Remove it
	// UNCOMMENT FOR TEST PURPOSES
	//appDetails := graphql.ApplicationExt{
	//	Application: graphql.Application{
	//		ID:   "7988ae0e-f008-4f62-aaeb-7c7cfb62a382",
	//		Name: "test2adas32321d",
	//	},
	//	Labels: map[string]interface{}{
	//		"compass/legacy-services": map[string]interface{}{"aa3b3ec4-3414-4b13-8314-7d26a4cf4356": map[string]interface{}{"id": "aa3b3ec4-3414-4b13-8314-7d26a4cf4356", "apiDefID": "df01eafc-a0d3-4cb2-8f58-be159fefde96", "eventDefID": "2d0a59b2-103d-4862-a240-4dce76124a87"}},
	//	},
	//}

	gqlCli := s.gqlProvider.GQLClient(r)
	gqlRequester := NewGqlRequester(gqlCli, &gql.Graphqlizer{}, &gql.GqlFieldsProvider{})
	labeler := NewAppLabeler()

	return NewServiceManager(gqlRequester, labeler, appDetails)
}
