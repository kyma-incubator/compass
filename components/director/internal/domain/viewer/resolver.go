package viewer

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// Resolver missing godoc
type Resolver struct {
}

// NewViewerResolver missing godoc
func NewViewerResolver() *Resolver {
	return &Resolver{}
}

// Viewer missing godoc
func (r *Resolver) Viewer(ctx context.Context) (*graphql.Viewer, error) {
	cons, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while getting viewer from context")
	}

	return ToViewer(cons)
}
