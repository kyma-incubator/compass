package service

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/pkg/errors"
)

type serviceManagerFactory struct {
	graphqlizer       *gql.Graphqlizer
	gqlFieldsProvider *gql.GqlFieldsProvider
}

func NewServiceManagerFactory() *serviceManagerFactory {
	return &serviceManagerFactory{
		graphqlizer:       &gql.Graphqlizer{},
		gqlFieldsProvider: &gql.GqlFieldsProvider{},
	}
}

func (s *serviceManagerFactory) ForRequest(r *http.Request) (ServiceManager, error) {
	appDetails, err := appdetails.LoadFromContext(r.Context())
	if err != nil {
		return nil, errors.Wrap(err, "while loading Application details from context")
	}

	gqlCli, err := gqlcli.LoadFromContext(r.Context())
	if err != nil {
		return nil, errors.Wrap(err, "while loading GraphQL client from context")
	}

	directorClient := director.NewClient(gqlCli, s.graphqlizer, s.gqlFieldsProvider)
	labeler := NewAppLabeler()

	return NewServiceManager(directorClient, labeler, appDetails)
}
