package inputvalidation

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type Validatable interface {
	Validate() error
}

type directive struct{}

func NewDirective() *directive {
	return &directive{}
}

func (d *directive) Validate(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	constructedObj, err := next(ctx)
	if err != nil {
		return nil, err
	}

	validatableObj, ok := constructedObj.(Validatable)
	if !ok {
		return nil, apperrors.NewInternalError(fmt.Sprintf("misuse of directive, object is not validatable: %T", constructedObj))
	}

	return validatableObj, Validate(validatableObj)
}
