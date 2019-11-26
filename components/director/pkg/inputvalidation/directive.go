package inputvalidation

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

type Validatable interface {
	Validate() error
}

type directive struct{}

func NewDirective() *directive {
	return &directive{}
}

func (d *directive) Validate(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	constructedObj, err := next(ctx)
	if err != nil {
		return nil, err
	}

	validatableObj, ok := constructedObj.(Validatable)
	if !ok {
		return nil, errors.Errorf("misuse of directive, object is not validatable: %T", constructedObj)

	}

	return validatableObj, validatableObj.Validate()
}
