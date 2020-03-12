package service

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	graphqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/pkg/errors"
)

type RequestContext struct {
	AppID          string
	DirectorClient DirectorClient
}

type requestContextProvider struct {
	graphqlizer       *graphqlizer.Graphqlizer
	gqlFieldsProvider *graphqlizer.GqlFieldsProvider
}

func NewRequestContextProvider() *requestContextProvider {
	return &requestContextProvider{
		graphqlizer:       &graphqlizer.Graphqlizer{},
		gqlFieldsProvider: &graphqlizer.GqlFieldsProvider{},
	}
}

func (s *requestContextProvider) ForRequest(r *http.Request) (RequestContext, error) {
	appDetails, err := appdetails.LoadFromContext(r.Context())
	if err != nil {
		return RequestContext{}, errors.Wrap(err, "while loading Application details from context")
	}

	gqlCli, err := gqlcli.LoadFromContext(r.Context())
	if err != nil {
		return RequestContext{}, errors.Wrap(err, "while loading GraphQL client from context")
	}

	directorClient := director.NewClient(gqlCli, s.graphqlizer, s.gqlFieldsProvider)

	return RequestContext{
		AppID:          appDetails.ID,
		DirectorClient: directorClient,
	}, nil
}
