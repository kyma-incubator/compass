package service

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) DetailsToGraphQL(in model.ServiceDetails) graphql.Application {
	panic("not implemented")
}

func (c *converter) GraphQLToDetailsModel(in graphql.Application) model.ServiceDetails {
	panic("not implemented")
}

func (c *converter) GraphQLToModel(in graphql.Application) model.Service {
	panic("not implemented")
}
