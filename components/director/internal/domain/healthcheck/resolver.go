package healthcheck

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=HealthCheckService -output=automock -outpkg=automock -case=underscore
type HealthCheckService interface{}

//go:generate mockery -name=HealthCheckConverter -output=automock -outpkg=automock -case=underscore
type HealthCheckConverter interface{}

type Resolver struct {
	svc       HealthCheckService
	converter HealthCheckConverter
}

func NewResolver(svc HealthCheckService) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &converter{},
	}
}

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
