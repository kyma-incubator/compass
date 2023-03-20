package portapiref

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const portApiRefTable string = `"public"."port_api_reference"`

var (
	portApiRefColumns = []string{"id", "app_id", "port_id", "api_id"}
)

type pgRepository struct {
	creator repo.Creator
}

// NewRepository returns a new entity responsible for repo-layer PorApiRef operations.
func NewRepository() *pgRepository {
	return &pgRepository{
		creator: repo.NewCreator(portApiRefTable, portApiRefColumns),
	}
}

// PortApiRefCollection is an array of Entities
type PortApiRefCollection []Entity

// Len returns the length of the collection
func (r PortApiRefCollection) Len() int {
	return len(r)
}

// Create creates a PortApiRef.
func (r *pgRepository) Create(ctx context.Context, tenant string, id, appID, portID, apiID string) error {
	entity := &Entity{
		ID:            id,
		ApplicationID: appID,
		PortID:        portID,
		ApiID:         apiID,
	}
	err := r.creator.Create(ctx, resource.PortApiRef, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}
