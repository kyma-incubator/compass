package api

import (
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/internalschema"
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

type InternalResolver struct {
	TokenResolver
}

type internalMutationResolver struct {
	*InternalResolver
}

type internalQueryResolver struct {
	*InternalResolver
}

func (r *InternalResolver) Mutation() internalschema.MutationResolver {
	return &internalMutationResolver{r}
}
func (r *InternalResolver) Query() internalschema.QueryResolver {
	return &internalQueryResolver{r}
}
