package token

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=TokenService -output=automock -outpkg=automock -case=underscore
type TokenService interface {
	GenerateOneTimeToken(ctx context.Context, runtimeID string, tokenType TokenType) (model.OneTimeToken, error)
}

//go:generate mockery -name=OneTimeTokenConverter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToGraphQL(model model.OneTimeToken) (graphql.OneTimeToken, error)
}

type Resolver struct {
	transact persistence.Transactioner
	svc      TokenService
	c        Converter
}

func NewTokenResolver(transact persistence.Transactioner, svc TokenService) *Resolver {
	return &Resolver{transact: transact, svc: svc}
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

	//TODO: Write converter
	return &graphql.OneTimeToken{Token: token.Token, ConnectorURL: token.ConnectorURL}, nil
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

	//TODO: Write converter
	return &graphql.OneTimeToken{Token: token.Token, ConnectorURL: token.ConnectorURL}, nil
}
