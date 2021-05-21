package api

import (
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

type ExternalResolver struct {
	CertificateResolver
}

type externalMutationResolver struct {
	*ExternalResolver
}

type externalQueryResolver struct {
	*ExternalResolver
}

func (r *ExternalResolver) Mutation() externalschema.MutationResolver {
	return &externalMutationResolver{r}
}
func (r *ExternalResolver) Query() externalschema.QueryResolver {
	return &externalQueryResolver{r}
}
