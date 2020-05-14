package customerrors

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	if val, ok := err.(GraphqlError); ok {
		gqlErr := &gqlerror.Error{
			Message:    err.Error(),
			Path:       graphql.GetResolverContext(ctx).Path(),
			Extensions: map[string]interface{}{"status_code": val.StatusCode},
		}
		return gqlErr
	}
	return graphql.DefaultErrorPresenter(ctx, err)
}
