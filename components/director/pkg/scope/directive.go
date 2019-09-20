package scope

import (
	"context"
	"github.com/pkg/errors"
)
import "github.com/99designs/gqlgen/graphql"

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

func (d *directive) Has(ctx context.Context, obj interface{}, next graphql.Resolver, scopesDefinition string) (res interface{}, err error) {
	actualScopes, err := LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	requiredScopes, err := d.scopesGetter.GetRequiredScopes(scopesDefinition)
	if err != nil {
		return nil, errors.Wrap(err, "while getting required scopes")
	}

	if d.matches(actualScopes, requiredScopes) {
		return next(ctx)
	}
	return nil, InsufficientScopesError
}

func (d *directive) matches(actual []string, required []string) bool {
	actMap := make(map[string]interface{})

	for _, a := range actual {
		actMap[a] = struct{}{}
	}
	for _, r := range required {
		_, ex := actMap[r]
		if !ex {
			return false
		}
	}
	return true
}
