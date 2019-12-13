package onetimetoken

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=TokenService -output=automock -outpkg=automock -case=underscore
type TokenService interface {
	GenerateOneTimeToken(ctx context.Context, runtimeID string, tokenType model.SystemAuthReferenceObjectType) (model.OneTimeToken, error)
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

func (r *Resolver) RequestOneTimeTokenForRuntime(ctx context.Context, id string) (*graphql.OneTimeToken, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	token, err := r.svc.GenerateOneTimeToken(ctx, id, model.RuntimeReference)
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

func (r *Resolver) RequestOneTimeTokenForApplication(ctx context.Context, id string) (*graphql.OneTimeToken, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	token, err := r.svc.GenerateOneTimeToken(ctx, id, model.ApplicationReference)
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

func (r *Resolver) RawEncoded(ctx context.Context, obj *graphql.OneTimeToken) (string, error) {
	if obj == nil {
		return "", errors.New("Token was nil")
	}

	rawJson, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	rawBaseEncoded := base64.StdEncoding.EncodeToString(rawJson)

	return rawBaseEncoded, nil
}

func (r *Resolver) Raw(ctx context.Context, obj *graphql.OneTimeToken) (string, error) {
	if obj == nil {
		return "", errors.New("Token was nil")
	}

	rawJson, err := json.Marshal(obj)

	if err != nil {
		return "", err
	}

	return string(rawJson), nil
}
