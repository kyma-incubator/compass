package token

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type OneTimeTokenConverter struct {
}

func (c OneTimeTokenConverter) ToGraphQL(model model.OneTimeToken) (graphql.OneTimeToken) {
	return graphql.OneTimeToken{Token: model.Token, ConnectorURL: model.ConnectorURL}
}
