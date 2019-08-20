package api

import (
	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type Resolver struct {
	TokenResolver
	CertificateResolver
}

type mutationResolver struct {
	*Resolver
}

type queryResolver struct {
	*Resolver
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &queryResolver{r}
}
