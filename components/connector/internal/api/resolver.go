package api

import (
	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type Resolver struct {
	CertificateResolver
	TokenResolver
}

type externalMutationResolver struct {
	*Resolver
}

type externalQueryResolver struct {
	*Resolver
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &externalMutationResolver{r}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &externalQueryResolver{r}
}
