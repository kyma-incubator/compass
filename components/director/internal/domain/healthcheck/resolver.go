package healthcheck

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// HealthCheckService missing godoc
//go:generate mockery --name=HealthCheckService --output=automock --outpkg=automock --case=underscore --disable-version-string
type HealthCheckService interface{}

// HealthCheckConverter missing godoc
//go:generate mockery --name=HealthCheckConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type HealthCheckConverter interface{}

// Resolver missing godoc
type Resolver struct {
	svc       HealthCheckService
	converter HealthCheckConverter
}

// NewResolver missing godoc
func NewResolver(svc HealthCheckService) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &converter{},
	}
}

// HealthChecks missing godoc
func (r *Resolver) HealthChecks(ctx context.Context, types []graphql.HealthCheckType, origin *string, first *int, after *graphql.PageCursor) (*graphql.HealthCheckPage, error) {
	return &graphql.HealthCheckPage{
		Data: []*graphql.HealthCheck{},
		PageInfo: &graphql.PageInfo{
			HasNextPage: false,
			EndCursor:   "",
			StartCursor: "",
		},
		TotalCount: 0,
	}, nil
}
