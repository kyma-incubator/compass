package customerrors

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

func Resolve(ctx context.Context, errID int) (*graphql.CustomError, error) {
	var err error
	var val *graphql.CustomError = nil
	switch errID {
	case 1:
		{
			err = NewErrorBuilder(NotUnique).With("damian", "testuje").Build()
			err = errors.Wrap(err, "while doing 1st thing")
			err = errors.Wrap(err, "while doing 2st thing")
		}
	case 2:
		{
			err = errors.New("something wrong")
		}
	case 3:
		val = &graphql.CustomError{
			Message: "To jest odpowiedz",
		}
	}

	return val, err
}
