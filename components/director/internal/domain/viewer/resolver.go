package viewer

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct {
}

func (r *Resolver) Viewer(ctx context.Context) (*graphql.ViewerInfo, error) {
	info := graphql.ViewerInfo{}
	info.ID = LoadIDFromContext(ctx)
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	info.Tenant = &tnt
	typeFromCtx := LoadTypeFromContext(ctx)
	var viewerType graphql.ViewerType
	switch typeFromCtx {
	case "Runtime":
		viewerType = graphql.ViewerTypeRuntime
	case "Application":
		viewerType = graphql.ViewerTypeApplication
	case "Integration System":
		viewerType = graphql.ViewerTypeIntegrationSystem
	case "Static User":
		viewerType = graphql.ViewerTypeUser
	}
	info.ViewerType = viewerType

	return &info, nil
}
