package onetimetoken

import (
	"fmt"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
	legacyConnectorURL string
}

func NewConverter(legacyConnectorURL string) *converter {
	return &converter{legacyConnectorURL}
}

func (c converter) ToGraphQLForRuntime(model model.OneTimeToken) graphql.OneTimeTokenForRuntime {
	return graphql.OneTimeTokenForRuntime{
		TokenWithURL: graphql.TokenWithURL{
			Token:        model.Token,
			ConnectorURL: model.ConnectorURL,
		},
	}
}

func (c converter) ToGraphQLForApplication(model model.OneTimeToken) (graphql.OneTimeTokenForApplication, error) {
	legacyConnectorURL, err := url.Parse(c.legacyConnectorURL)
	if err != nil {
		return graphql.OneTimeTokenForApplication{}, errors.Wrapf(err, "while parsing string (%s) as the URL", c.legacyConnectorURL)
	}

	if legacyConnectorURL.RawQuery != "" {
		legacyConnectorURL.RawQuery += "&"
	}
	legacyConnectorURL.RawQuery += fmt.Sprintf("token=%s", model.Token)

	return graphql.OneTimeTokenForApplication{
		TokenWithURL: graphql.TokenWithURL{
			Token:        model.Token,
			ConnectorURL: model.ConnectorURL,
		},
		LegacyConnectorURL: legacyConnectorURL.String(),
	}, nil
}
