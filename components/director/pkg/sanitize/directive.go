package sanitize

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

//go:generate mockery -name=ScopesGetter -output=automock -outpkg=automock -case=underscore
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}

type directive struct {
	scopesGetter ScopesGetter
}

func NewDirective(getter ScopesGetter) *directive {
	return &directive{
		scopesGetter: getter,
	}
}

func (d *directive) Sanitize(ctx context.Context, obj interface{}, next graphql.Resolver, scopesDefinition string) (interface{}, error) {
	actualScopes, err := scope.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	requiredScopes, err := d.scopesGetter.GetRequiredScopes(scopesDefinition)
	if err != nil {
		log.C(ctx).Warnf("Stripping sensitive data from %T", obj)
		return nil, nil
	}

	if !str.Matches(actualScopes, requiredScopes) {
		return nil, apperrors.NewInsufficientScopesError(requiredScopes, actualScopes)
	}
	return next(ctx)
}
