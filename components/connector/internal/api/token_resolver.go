package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
)

//go:generate mockery -name=TokenResolver
type TokenResolver interface {
	GenerateApplicationToken(ctx context.Context, appID string) (*externalschema.Token, error)
	GenerateRuntimeToken(ctx context.Context, runtimeID string) (*externalschema.Token, error)
	IsHealthy(ctx context.Context) (bool, error)
}

type tokenResolver struct {
	tokenService tokens.Service
}

func NewTokenResolver(tokenService tokens.Service) TokenResolver {
	return &tokenResolver{
		tokenService: tokenService,
	}
}

func (r *tokenResolver) GenerateApplicationToken(ctx context.Context, appID string) (*externalschema.Token, error) {
	token, err := r.tokenService.CreateToken(appID, tokens.ApplicationToken)
	if err != nil {
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create Application token")
	}

	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*externalschema.Token, error) {
	token, err := r.tokenService.CreateToken(runtimeID, tokens.RuntimeToken)
	if err != nil {
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create Runtime token")
	}

	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) IsHealthy(ctx context.Context) (bool, error) {
	return true, nil
}
