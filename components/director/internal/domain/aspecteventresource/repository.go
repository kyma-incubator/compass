package aspecteventresource

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	aspectEventResourcesTable string = `public.aspect_event_resources`
	aspectIDColumn            string = "aspect_id"
	appIDColumn               string = "app_id"
)

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
	creator     repo.Creator
	deleter     repo.Deleter
	unionLister repo.UnionLister
	lister      repo.Lister

	conv AspectEventResourceConverter
}

// NewRepository returns a new entity responsible for repo-layer Aspect Event Resource operations.
func NewRepository(conv AspectEventResourceConverter) *pgRepository {
	return &pgRepository{
		creator:     repo.NewCreator(aspectEventResourcesTable, aspectEventResourcesColumns),
		deleter:     repo.NewDeleter(aspectEventResourcesTable),
		unionLister: repo.NewUnionLister(aspectEventResourcesTable, aspectEventResourcesColumns),
		lister:      repo.NewLister(aspectEventResourcesTable, aspectEventResourcesColumns),

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

// ListByAspectID lists Aspect Event Resources by Aspect id
func (r *pgRepository) ListByAspectID(ctx context.Context, tenant, aspectID string) ([]*model.AspectEventResource, error) {
	aspectEventResourceCollection := AspectEventResourceCollection{}

	condition := repo.NewEqualCondition(aspectIDColumn, aspectID)

	if err := r.lister.List(ctx, resource.AspectEventResource, tenant, &aspectEventResourceCollection, condition); err != nil {
		return nil, err
	}

	aspectEventResources := make([]*model.AspectEventResource, 0, aspectEventResourceCollection.Len())
	for _, aspectEventResource := range aspectEventResourceCollection {
		aspectEventResourceModel := r.conv.FromEntity(&aspectEventResource)
		aspectEventResources = append(aspectEventResources, aspectEventResourceModel)
	}

	return aspectEventResources, nil
}

// ListByApplicationIDs retrieves all Aspect Event Resources matching an array of applicationIDs from the Compass storage.
func (r *pgRepository) ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.AspectEventResource, map[string]int, error) {
	aspectEventResourceCollection := AspectEventResourceCollection{}
	orderByColumns := repo.OrderByParams{repo.NewAscOrderBy(aspectIDColumn), repo.NewAscOrderBy(appIDColumn)}

	counts, err := r.unionLister.List(ctx, resource.AspectEventResource, tenantID, applicationIDs, appIDColumn, pageSize, cursor, orderByColumns, &aspectEventResourceCollection)
	if err != nil {
		return nil, nil, err
	}

	aspectEventResources := make([]*model.AspectEventResource, 0, len(aspectEventResourceCollection))
	for _, d := range aspectEventResourceCollection {
		entity := r.conv.FromEntity(&d)

		aspectEventResources = append(aspectEventResources, entity)
	}

	return aspectEventResources, counts, nil
}
