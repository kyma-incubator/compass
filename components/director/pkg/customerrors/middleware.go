package customerrors

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
)

func HandlerErrors(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	val, err := next(ctx)
	if err != nil {
		customErr := &Error{}
		found := errors.As(err, customErr)

		if !found {
			//TODO: Use Logger
			fmt.Printf("Not handled error yet: %v\n", err)
			return val, err
		} else {
			if customErr.errorCode == InternalError {
				fmt.Printf("Internal Error: %v\n", err.Error())
				return val, GraphqlError{
					StatusCode: InternalError,
					Message:    "Internal error in director",
				}
			}

			graphqlErr := GraphqlError{
				StatusCode: customErr.errorCode,
				Message:    customErr.Error(),
			}
			return val, graphqlErr
		}
	}
	return val, err
}
