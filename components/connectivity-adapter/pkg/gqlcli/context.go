package gqlcli

import (
	"context"
	"errors"
)

type GraphQLClientContextKey struct{}

var NoGraphQLClientError = errors.New("cannot read Application details from context")
var NilContextError = errors.New("context is empty")

func LoadFromContext(ctx context.Context) (GraphQLClient, error) {
	if ctx == nil {
		return nil, NilContextError
	}

	value := ctx.Value(GraphQLClientContextKey{})

	cli, ok := value.(GraphQLClient)

	if !ok {
		return nil, NoGraphQLClientError
	}

	return cli, nil
}

func SaveToContext(ctx context.Context, cli GraphQLClient) context.Context {
	return context.WithValue(ctx, GraphQLClientContextKey{}, cli)
}
