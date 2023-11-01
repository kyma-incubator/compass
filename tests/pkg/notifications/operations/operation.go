package operations

import (
	"context"
	"testing"

	gcli "github.com/machinebox/graphql"
)

type Operation interface {
	Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client)
	Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client)
}
