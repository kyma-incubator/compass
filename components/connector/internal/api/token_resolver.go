package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type TokenResolver interface {
	GenerateApplicationToken(ctx context.Context, systemAuthID string) (*externalschema.Token, error)
	GenerateRuntimeToken(ctx context.Context, systemAuthID string) (*externalschema.Token, error)
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

func (r *tokenResolver) GenerateApplicationToken(_ context.Context, systemAuthID string) (*externalschema.Token, error) {
	r.log.Infof("Generating one-time token for Application with systemAuthID %s", systemAuthID)

	token, err := r.tokenService.CreateToken(systemAuthID, tokens.ApplicationToken)
	if err != nil {
		r.log.Errorf("Error occurred while creating one-time token for Application with systemAuthID %s : %s ", systemAuthID, err.Error())
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create one-time token for Application")
	}

	r.log.Infof("One-time token generated successfully for Application with systemAuthID %s", systemAuthID)
	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) GenerateRuntimeToken(_ context.Context, systemAuthID string) (*externalschema.Token, error) {
	r.log.Infof("Generating one-time token for Runtime with systemAuthID %s", systemAuthID)

	token, err := r.tokenService.CreateToken(systemAuthID, tokens.RuntimeToken)
	if err != nil {
		r.log.Errorf("Error occurred while creating one-time token for Runtime with systemAuthID %s : %s ", systemAuthID, err.Error())
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create one-time token for Runtime")
	}

	r.log.Infof("One-time token generated successfully for Runtime with systemAuthID %s", systemAuthID)
	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) IsHealthy(_ context.Context) (bool, error) {
	return true, nil
}
