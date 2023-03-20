package port

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const portsTable string = `"public"."ports"`

var (
	idColumn    = "id"
	portColumns = []string{"id", "data_product_id", "app_id", "name", "port_type", "description", "producer_cardinality", "disabled"}
	idColumns   = []string{"id"}
)

// PortConverter converts Ports between the model.Port service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=PortConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type PortConverter interface {
	FromEntity(entity *Entity) *model.Port
	ToEntity(apiModel *model.Port) *Entity
}

type pgRepository struct {
	creator repo.Creator
	conv    PortConverter
}

// NewRepository returns a new entity responsible for repo-layer Port operations.
func NewRepository(conv PortConverter) *pgRepository {
	return &pgRepository{
		creator: repo.NewCreator(portsTable, portColumns),
		conv:    conv,
	}
}

// PortCollection is an array of Entities
type PortCollection []Entity

// Len returns the length of the collection
func (r PortCollection) Len() int {
	return len(r)
}

// Create creates a Port.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.Port) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creator.Create(ctx, resource.Port, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}
