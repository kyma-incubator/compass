package healthcheck

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

type svc interface{}

type Resolver struct {
	svc       svc
	converter *Converter
}

func NewResolver(svc svc) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &Converter{},
	}
}

func (r *Resolver) HealthChecks(ctx context.Context, types []graphql.HealthCheckType, origin *string, first *int, after *string) (*graphql.HealthCheckPage, error) {
	panic("not implemented")
}
