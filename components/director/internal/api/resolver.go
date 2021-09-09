package api

import "github.com/kyma-incubator/compass/components/director/pkg/graphql/internalschema"

// InternalResolver missing godoc
type InternalResolver struct {
	TokenResolver
}

type internalMutationResolver struct {
	*InternalResolver
}

type internalQueryResolver struct {
	*InternalResolver
}

// Mutation missing godoc
func (r *InternalResolver) Mutation() internalschema.MutationResolver {
	return &internalMutationResolver{r}
}

// Query missing godoc
func (r *InternalResolver) Query() internalschema.QueryResolver {
	return &internalQueryResolver{r}
}
