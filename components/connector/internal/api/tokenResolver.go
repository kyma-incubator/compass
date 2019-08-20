package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type TokenResolver struct {
}

func (r *TokenResolver) GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
func (r *TokenResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
