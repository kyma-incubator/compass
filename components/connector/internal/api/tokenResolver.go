package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

//go:generate mockery -name=TokenResolver
type TokenResolver interface {
	GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error)
	GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error)
}

type tokenResolver struct {
}

func NewTokenResolver() TokenResolver {
	return &tokenResolver{}
}

func (r *tokenResolver) GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
func (r *tokenResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
