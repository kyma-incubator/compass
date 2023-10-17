package resource_providers

import (
	"context"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type Provider interface {
	Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string
	Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client)
}
