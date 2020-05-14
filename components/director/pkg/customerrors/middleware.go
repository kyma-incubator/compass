package customerrors

import (
	"context"
	"errors"
	"fmt"
	gqgraphql "github.com/99designs/gqlgen/graphql"
)

func HandlerErrors(ctx context.Context, next gqgraphql.Resolver) (res interface{}, err error) {
	//TODO: Tutaj zapnij mapowanie errorów
	val, err := next(ctx)
	if err != nil {
		//TODO: Map this error
		fmt.Printf("Found error: %s", err.Error())


		if ok := errors.Is(err, InternalErr); ok {
			fmt.Printf("Internal Error: %v\n", err.Error())
			return val, GraphqlError{
				StatusCode: InternalError,
				Message:    "Internal error in director",
			}
		}

		var customErr Error
		for ; ; {
			val, ok := err.(Error)
			if ok {
				customErr = val
				break
			}
			if err = errors.Unwrap(err); err == nil {
				break
			}

		}
		graphqlErr := GraphqlError{
			StatusCode: customErr.statusCode,
			Message:    customErr.Error(),
		}

		return val, graphqlErr
	}
	return val, err
}
