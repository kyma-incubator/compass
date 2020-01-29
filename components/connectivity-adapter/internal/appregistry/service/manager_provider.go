package service

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/pkg/errors"
)

type serviceManagerProvider struct {
	graphqlizer       *gql.Graphqlizer
	gqlFieldsProvider *gql.GqlFieldsProvider
}

func NewServiceManagerProvider() *serviceManagerProvider {
	return &serviceManagerProvider{
		graphqlizer:       &gql.Graphqlizer{},
		gqlFieldsProvider: &gql.GqlFieldsProvider{},
	}
}

func (s *serviceManagerProvider) ForRequest(r *http.Request) (ServiceManager, error) {
	appDetails, err := appdetails.LoadFromContext(r.Context())
	if err != nil {
		return nil, errors.Wrap(err, "while loading Application details from context")
	}

	gqlCli, err := gqlcli.LoadFromContext(r.Context())
	if err != nil {
		return nil, errors.Wrap(err, "while loading GraphQL client from context")
	}

	gqlRequester := NewGqlRequester(gqlCli, s.graphqlizer, s.gqlFieldsProvider)
	labeler := NewAppLabeler()

	return NewServiceManager(gqlRequester, labeler, appDetails)
}
