package paginator

import (
	"context"

	gcli "github.com/machinebox/graphql"
)

//go:generate mockery -name=Client
type Client interface {
	Do(ctx context.Context, req *gcli.Request, res interface{}) error
}
