package resource_providers

import (
	"context"
	"testing"

	gcli "github.com/machinebox/graphql"
)

type Provider interface {
	Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string
	Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client)
}
