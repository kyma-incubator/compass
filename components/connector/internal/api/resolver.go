package api

import (
	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type Resolver struct {
}

type mutationResolver struct {
	*Resolver
	TokenResolver
	CertificateResolver
}

type queryResolver struct {
	*Resolver
	CertificateResolver
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &mutationResolver{r, TokenResolver{}, CertificateResolver{}}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &queryResolver{r, CertificateResolver{}}
}
