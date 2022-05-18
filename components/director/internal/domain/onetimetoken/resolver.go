package onetimetoken

import (
	"context"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// TokenService missing godoc
//go:generate mockery --name=TokenService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TokenService interface {
	GenerateOneTimeToken(ctx context.Context, runtimeID string, tokenType pkgmodel.SystemAuthReferenceObjectType) (*model.OneTimeToken, error)
	RegenerateOneTimeToken(ctx context.Context, sysAuthID string) (*model.OneTimeToken, error)
}

// TokenConverter missing godoc
//go:generate mockery --name=TokenConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type TokenConverter interface {
	ToGraphQLForRuntime(model model.OneTimeToken) graphql.OneTimeTokenForRuntime
	ToGraphQLForApplication(model model.OneTimeToken) (graphql.OneTimeTokenForApplication, error)
}

// Resolver missing godoc
type Resolver struct {
	transact              persistence.Transactioner
	svc                   TokenService
	conv                  TokenConverter
	suggestTokenHeaderKey string
}

// NewTokenResolver missing godoc
func NewTokenResolver(transact persistence.Transactioner, svc TokenService, conv TokenConverter, suggestTokenHeaderKey string) *Resolver {
	return &Resolver{transact: transact, svc: svc, conv: conv, suggestTokenHeaderKey: suggestTokenHeaderKey}
}

// RequestOneTimeTokenForRuntime missing godoc
func (r *Resolver) RequestOneTimeTokenForRuntime(ctx context.Context, id string, systemAuthID *string) (*graphql.OneTimeTokenForRuntime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var token *model.OneTimeToken
	if systemAuthID != nil {
		token, err = r.svc.RegenerateOneTimeToken(ctx, *systemAuthID)
	} else {
		token, err = r.svc.GenerateOneTimeToken(ctx, id, pkgmodel.RuntimeReference)
	}
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gqlToken := r.conv.ToGraphQLForRuntime(*token)
	return &gqlToken, nil
}

// RequestOneTimeTokenForApplication missing godoc
func (r *Resolver) RequestOneTimeTokenForApplication(ctx context.Context, id string, systemAuthID *string) (*graphql.OneTimeTokenForApplication, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var token *model.OneTimeToken
	if systemAuthID != nil {
		token, err = r.svc.RegenerateOneTimeToken(ctx, *systemAuthID)
	} else {
		token, err = r.svc.GenerateOneTimeToken(ctx, id, pkgmodel.ApplicationReference)
	}
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

// RawEncoded missing godoc
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
		Used:         obj.Used,
		ExpiresAt:    obj.ExpiresAt,
		CreatedAt:    obj.CreatedAt,
		UsedAt:       obj.UsedAt,
		Type:         obj.Type,
	})
}

// Raw missing godoc
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
		Used:         obj.Used,
		ExpiresAt:    obj.ExpiresAt,
		CreatedAt:    obj.CreatedAt,
		UsedAt:       obj.UsedAt,
		Type:         obj.Type,
	})
}
