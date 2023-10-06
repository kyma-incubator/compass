package operations

import (
	"context"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type Operation interface {
	Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client)
	Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client)
}
