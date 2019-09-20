package token

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=TokenService -output=automock -outpkg=automock -case=underscore
type TokenService interface {
	GenerateOneTimeToken(ctx context.Context, runtimeID string, tokenType Type) (model.OneTimeToken, error)
}

//go:generate mockery -name=TokenConverter -output=automock -outpkg=automock -case=underscore
type TokenConverter interface {
	ToGraphQL(model model.OneTimeToken) graphql.OneTimeToken
}

type Resolver struct {
	transact persistence.Transactioner
	svc      TokenService
	conv     TokenConverter
}

func NewTokenResolver(transact persistence.Transactioner, svc TokenService, conv TokenConverter) *Resolver {
	return &Resolver{transact: transact, svc: svc, conv: conv}
}

func (r *Resolver) GenerateOneTimeTokenForRuntime(ctx context.Context, id string) (*graphql.OneTimeToken, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	token, err := r.svc.GenerateOneTimeToken(ctx, id, RuntimeToken)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while commiting transaction")
	}

	gqlToken := r.conv.ToGraphQL(token)
	return &gqlToken, nil
}

func (r *Resolver) GenerateOneTimeTokenForApp(ctx context.Context, id string) (*graphql.OneTimeToken, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	token, err := r.svc.GenerateOneTimeToken(ctx, id, ApplicationToken)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while commiting transaction")
	}
	gqlToken := r.conv.ToGraphQL(token)
	return &gqlToken, nil
}
