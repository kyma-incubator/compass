package api

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

//go:generate mockery -name=TokenResolver
type TokenResolver interface {
	GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error)
	GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error)
}

type tokenResolver struct {
	tokenService tokens.Service
}

func NewTokenResolver(tokenService tokens.Service) TokenResolver {
	return &tokenResolver{
		tokenService: tokenService,
	}
}

func (r *tokenResolver) GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error) {
	token, err := r.tokenService.CreateToken(appID, tokens.ApplicationToken)
	if err != nil {
		return &gqlschema.Token{}, errors.Wrap(err, "Failed to create Application token")
	}

	return &gqlschema.Token{Token: token}, nil
}

func (r *tokenResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error) {
	token, err := r.tokenService.CreateToken(runtimeID, tokens.RuntimeToken)
	if err != nil {
		return &gqlschema.Token{}, errors.Wrap(err, "Failed to create Runtime token")
	}

	return &gqlschema.Token{Token: token}, nil
}
