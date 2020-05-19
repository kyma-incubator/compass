package customerrors

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	var graphqlErr GraphqlError
	if found := errors.As(err, &graphqlErr); found {
		gqlErr := &gqlerror.Error{
			Message:    err.Error(),
			Path:       graphql.GetResolverContext(ctx).Path(),
			Extensions: map[string]interface{}{"error_code": graphqlErr.StatusCode},
		}
		return gqlErr
	}
	return graphql.DefaultErrorPresenter(ctx, err)
}
