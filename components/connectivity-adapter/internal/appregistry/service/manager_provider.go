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

	gqlCli := s.gqlProvider.GQLClient(r)
	gqlRequester := NewGqlRequester(gqlCli, &gql.Graphqlizer{}, &gql.GqlFieldsProvider{})
	labeler := NewAppLabeler()

	return NewServiceManager(gqlRequester, labeler, appDetails)
}
