package inputvalidation

import (
	"context"
	"fmt"
	"strings"

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

func (d *directive) Validate(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	constructedObj, err := next(ctx)
	if err != nil {
		return nil, err
	}

	validatableObj, ok := constructedObj.(Validatable)
	if !ok {
		return nil, errors.Errorf("misuse of directive, object is not validatable: %T", constructedObj)

	}

	var typeName string
	split := strings.Split(fmt.Sprintf("%T", constructedObj), ".")
	if len(split) > 1 {
		typeName = split[1]
	} else {
		typeName = split[0]
	}

	return validatableObj, errors.Wrapf(validatableObj.Validate(), "validation error for type %s", typeName)
}
