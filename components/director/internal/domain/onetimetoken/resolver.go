package onetimetoken

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery --name=TokenService --output=automock --outpkg=automock --case=underscore
type TokenService interface {
	GenerateOneTimeToken(ctx context.Context, runtimeID string, tokenType model.SystemAuthReferenceObjectType) (*model.OneTimeToken, error)
}

//go:generate mockery --name=TokenConverter --output=automock --outpkg=automock --case=underscore
type TokenConverter interface {
	ToGraphQLForRuntime(model model.OneTimeToken) graphql.OneTimeTokenForRuntime
	ToGraphQLForApplication(model model.OneTimeToken) (graphql.OneTimeTokenForApplication, error)
}

type Resolver struct {
	transact              persistence.Transactioner
	svc                   TokenService
	conv                  TokenConverter
	suggestTokenHeaderKey string
}

func NewTokenResolver(transact persistence.Transactioner, svc TokenService, conv TokenConverter, suggestTokenHeaderKey string) *Resolver {
	return &Resolver{transact: transact, svc: svc, conv: conv, suggestTokenHeaderKey: suggestTokenHeaderKey}
}

func (r *Resolver) RequestOneTimeTokenForRuntime(ctx context.Context, id string) (*graphql.OneTimeTokenForRuntime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	token, err := r.svc.GenerateOneTimeToken(ctx, id, model.RuntimeReference)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gqlToken := r.conv.ToGraphQLForRuntime(*token)
	return &gqlToken, nil
}

func (r *Resolver) RequestOneTimeTokenForApplication(ctx context.Context, id string) (*graphql.OneTimeTokenForApplication, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	token, err := r.svc.GenerateOneTimeToken(ctx, id, model.ApplicationReference)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}
	gqlToken, err := r.conv.ToGraphQLForApplication(*token)
	if err != nil {
		return nil, errors.Wrap(err, "while converting one-time token to graphql")
	}
	return &gqlToken, nil
}

func (r *Resolver) RawEncoded(ctx context.Context, obj *graphql.TokenWithURL) (*string, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Token was nil")
	}

	if !tokenSuggestionEnabled(ctx, r.suggestTokenHeaderKey) {
		return rawEncoded(obj)
	}

	return rawEncoded(&graphql.TokenWithURL{
		Token:        extractToken(obj.Token),
		ConnectorURL: obj.ConnectorURL,
	})
}

func (r *Resolver) Raw(ctx context.Context, obj *graphql.TokenWithURL) (*string, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Token was nil")
	}

	if !tokenSuggestionEnabled(ctx, r.suggestTokenHeaderKey) {
		return raw(obj)
	}

	return raw(&graphql.TokenWithURL{
		Token:        extractToken(obj.Token),
		ConnectorURL: obj.ConnectorURL,
	})
}
