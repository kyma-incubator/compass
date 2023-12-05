package aspecteventresource

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const aspectEventResourcecTable string = `public.aspect_event_resources`

var (
	aspectEventResourcesColumns = []string{"id", "app_id", "app_template_version_id", "aspect_id", "ord_id", "min_version", "subset", "ready", "created_at", "updated_at", "deleted_at", "error"}
)

// AspectEventResourceConverter converts Aspect Event Resources between the model.AspectEventResource service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=AspectEventResourceConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectEventResourceConverter interface {
	FromEntity(entity *Entity) *model.AspectEventResource
	ToEntity(aspectEventResourceModel *model.AspectEventResource) *Entity
}

type pgRepository struct {
	creator repo.Creator
	deleter repo.Deleter

	conv AspectEventResourceConverter
}

// NewRepository returns a new entity responsible for repo-layer Aspect Event Resource operations.
func NewRepository(conv AspectEventResourceConverter) *pgRepository {
	return &pgRepository{
		creator: repo.NewCreator(aspectEventResourcecTable, aspectEventResourcesColumns),
		deleter: repo.NewDeleter(aspectEventResourcecTable),

		conv: conv,
	}
}

// AspectEventResourceCollection is an array of Entities
type AspectEventResourceCollection []Entity

// Len returns the length of the collection
func (r AspectEventResourceCollection) Len() int {
	return len(r)
}

// Create creates an Aspect Event Resource for an Aspect.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.AspectEventResource) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	err := r.creator.Create(ctx, resource.AspectEventResource, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}
