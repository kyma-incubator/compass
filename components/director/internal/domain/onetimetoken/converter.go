package onetimetoken

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c converter) ToGraphQL(model model.OneTimeToken) graphql.OneTimeToken {
	return graphql.OneTimeToken{Token: model.Token, ConnectorURL: model.ConnectorURL}
}
