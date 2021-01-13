package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/pkg/errors"
)

type TokenResolver interface {
	GenerateApplicationToken(ctx context.Context, authID string) (*externalschema.Token, error)
	GenerateRuntimeToken(ctx context.Context, authID string) (*externalschema.Token, error)
	IsHealthy(ctx context.Context) (bool, error)
}

type tokenResolver struct {
	tokenService tokens.Service
	transact     persistence.Transactioner
}

func NewTokenResolver(transact persistence.Transactioner, tokenService tokens.Service) TokenResolver {
	return &tokenResolver{
		tokenService: tokenService,
		transact:     transact,
	}
}

func (r *tokenResolver) GenerateApplicationToken(ctx context.Context, authID string) (*externalschema.Token, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Generating one-time token for Application with authID %s", authID)

	token, err := r.tokenService.CreateToken(ctx, authID, tokens.ApplicationToken)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while creating one-time token for Application with authID %s", authID)
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create one-time token for Application")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("One-time token generated successfully for Application with authID %s", authID)
	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) GenerateRuntimeToken(ctx context.Context, authID string) (*externalschema.Token, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Generating one-time token for Runtime with authID %s", authID)

	token, err := r.tokenService.CreateToken(ctx, authID, tokens.RuntimeToken)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while creating one-time token for Runtime with authID %s", authID)
		return &externalschema.Token{}, errors.Wrap(err, "Failed to create one-time token for Runtime")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("One-time token generated successfully for Runtime with authID %s", authID)
	return &externalschema.Token{Token: token}, nil
}

func (r *tokenResolver) IsHealthy(_ context.Context) (bool, error) {
	return true, nil
}
