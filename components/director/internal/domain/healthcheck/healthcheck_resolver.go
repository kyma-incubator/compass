package healthcheck

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type Resolver struct {
	svc       *Service
	converter *Converter
}

func NewResolver(svc *Service) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &Converter{},
	}
}

func (r *Resolver) HealthChecks(ctx context.Context, types []gqlschema.HealthCheckType, origin *string, first *int, after *string) (*gqlschema.HealthCheckPage, error) {
	panic("not implemented")
}
