package onetimetoken

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct {
	legacyConnectorURL string
}

// NewConverter missing godoc
func NewConverter(legacyConnectorURL string) *converter {
	return &converter{legacyConnectorURL}
}

// ToGraphQLForRuntime missing godoc
func (c converter) ToGraphQLForRuntime(model model.OneTimeToken) graphql.OneTimeTokenForRuntime {
	return graphql.OneTimeTokenForRuntime{
		TokenWithURL: graphql.TokenWithURL{
			Token:        model.Token,
			ConnectorURL: model.ConnectorURL,
		},
	}
}

// ToGraphQLForApplication missing godoc
func (c converter) ToGraphQLForApplication(model model.OneTimeToken) (graphql.OneTimeTokenForApplication, error) {
	urlWithToken, err := legacyConnectorURLWithToken(c.legacyConnectorURL, model.Token)
	if err != nil {
		return graphql.OneTimeTokenForApplication{}, err
	}

	return graphql.OneTimeTokenForApplication{
		TokenWithURL: graphql.TokenWithURL{
			Token:        model.Token,
			ConnectorURL: model.ConnectorURL,
		},
		LegacyConnectorURL: urlWithToken,
	}, nil
}
