package scope

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery --name=ScopesGetter --output=automock --outpkg=automock --case=underscore
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}
type ScopesMismatchErrorProvider interface {
	Error([]string, []string) error
}

type directive struct {
	scopesGetter  ScopesGetter
	errorProvider ScopesMismatchErrorProvider
}

func NewDirective(getter ScopesGetter, errorProvider ScopesMismatchErrorProvider) *directive {
	return &directive{
		scopesGetter:  getter,
		errorProvider: errorProvider,
	}
}

func (d *directive) VerifyScopes(ctx context.Context, _ interface{}, next graphql.Resolver, scopesDefinition string) (interface{}, error) {
	actualScopes, err := LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	requiredScopes, err := d.scopesGetter.GetRequiredScopes(scopesDefinition)
	if err != nil {
		return nil, errors.Wrap(err, "while getting required scopes")
	}

	if !str.Matches(actualScopes, requiredScopes) {
		return nil, d.errorProvider.Error(requiredScopes, actualScopes)
	}
	return next(ctx)
}
