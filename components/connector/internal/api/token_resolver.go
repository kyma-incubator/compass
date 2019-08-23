package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=TokenResolver
type TokenResolver interface {
	GenerateApplicationToken(ctx context.Context, appID string) (*externalschema.Token, error)
	GenerateRuntimeToken(ctx context.Context, runtimeID string) (*externalschema.Token, error)
	IsHealthy(ctx context.Context) (bool, error)
}

type tokenResolver struct {
	tokenService tokens.Service
	log          *logrus.Entry
}

func NewTokenResolver(tokenService tokens.Service) TokenResolver {
	return &tokenResolver{
		tokenService: tokenService,
		log:          logrus.WithField("Resolver", "Token"),
	}
}

func (r *tokenResolver) GenerateApplicationToken(ctx context.Context, appID string) (*externalschema.Token, error) {
	r.log.Infof("Generating token for %s Application...", appID)

	token, err := r.tokenService.CreateToken(appID, tokens.ApplicationToken)
	if err != nil {
		r.log.Error(err.Error())
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create Application token")
	}

	r.log.Infof("Token generated for %s Application...", appID)
	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*externalschema.Token, error) {
	r.log.Infof("Generating token for %s Runtime...", runtimeID)

	token, err := r.tokenService.CreateToken(runtimeID, tokens.RuntimeToken)
	if err != nil {
		r.log.Error(err.Error())
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create Runtime token")
	}

	r.log.Infof("Token generated for %s Runtime...", runtimeID)
	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) IsHealthy(ctx context.Context) (bool, error) {
	return true, nil
}
