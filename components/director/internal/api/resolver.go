package api

import "github.com/kyma-incubator/compass/components/director/pkg/graphql/internalschema"

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
