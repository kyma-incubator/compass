package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/internalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// TokenResolver missing godoc
type TokenResolver interface {
	GenerateCSRToken(ctx context.Context, authID string) (*internalschema.Token, error)
	IsHealthy(ctx context.Context) (bool, error)
}

// TokenService missing godoc
//go:generate mockery --name=TokenService --output=automock --outpkg=automock --case=underscore
type TokenService interface {
	RegenerateOneTimeToken(ctx context.Context, authID string, token tokens.TokenType) (model.OneTimeToken, error)
}

type tokenResolver struct {
	transact     persistence.Transactioner
	tokenService TokenService
}

// NewTokenResolver missing godoc
func NewTokenResolver(transact persistence.Transactioner, tokenService TokenService) TokenResolver {
	return &tokenResolver{
		tokenService: tokenService,
		transact:     transact,
	}
}

// GenerateCSRToken missing godoc
func (r *tokenResolver) GenerateCSRToken(ctx context.Context, authID string) (*internalschema.Token, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Generating one-time token for CSR with authID %s", authID)

	token, err := r.tokenService.RegenerateOneTimeToken(ctx, authID, tokens.CSRToken)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while creating one-time token for CSR with authID %s: %v", authID, err)
		return nil, errors.Wrap(err, "Failed to create one-time token for CSR")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	log.C(ctx).Infof("One-time token generated successfully for CSR with authID %s", authID)
	return &internalschema.Token{Token: token.Token}, nil
}

// IsHealthy missing godoc
func (r *tokenResolver) IsHealthy(_ context.Context) (bool, error) {
	return true, nil
}
