package paginator

import (
	"context"

	gcli "github.com/machinebox/graphql"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Client
type Client interface {
	Do(ctx context.Context, req *gcli.Request, res interface{}) error
}
