package resolver

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type healthcheckResolver struct {

}

func (r *healthcheckResolver) HealthChecks(ctx context.Context, types []gqlschema.HealthCheckType, origin *string, first *int, after *string) (*gqlschema.HealthCheckPage, error) {
	panic("not implemented")
}
