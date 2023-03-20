package porteventref

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const portEventRefTable string = `"public"."port_event_reference"`

var (
	portEventRefColumns = []string{"id", "app_id", "port_id", "event_id", "min_version"}
)

type pgRepository struct {
	creator repo.Creator
}

// NewRepository returns a new entity responsible for repo-layer PorEventRef operations.
func NewRepository() *pgRepository {
	return &pgRepository{
		creator: repo.NewCreator(portEventRefTable, portEventRefColumns),
	}
}

// PortEventRefCollection is an array of Entities
type PortEventRefCollection []Entity

// Len returns the length of the collection
func (r PortEventRefCollection) Len() int {
	return len(r)
}

// Create creates a PortApiRef.
func (r *pgRepository) Create(ctx context.Context, tenant string, id, appID, portID, eventID string, minVersion *string) error {
	entity := &Entity{
		ID:            id,
		ApplicationID: appID,
		PortID:        portID,
		EventID:       eventID,
		MinVersion:    repo.NewNullableString(minVersion),
	}
	err := r.creator.Create(ctx, resource.PortEventRef, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}
