package aspect

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const aspectsTable string = `"public"."aspects"`

var (
	aspectColumns = []string{"id", "integration_dependency_id", "title", "description", "mandatory", "support_multiple_providers",
		"api_resources", "event_resources", "ready", "created_at", "updated_at", "deleted_at", "error"}
)

// AspectConverter converts Aspects between the model.Aspect service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=AspectConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectConverter interface {
	FromEntity(entity *Entity) *model.Aspect
	ToEntity(integrationDependencyModel *model.Aspect) *Entity
}

type pgRepository struct {
	singleGetter repo.SingleGetter
	lister       repo.Lister
	creator      repo.Creator
	deleter      repo.Deleter

	conv AspectConverter
}

// NewRepository returns a new entity responsible for repo-layer Aspects operations.
func NewRepository(conv AspectConverter) *pgRepository {
	return &pgRepository{
		singleGetter: repo.NewSingleGetter(aspectsTable, aspectColumns),
		lister:       repo.NewLister(aspectsTable, aspectColumns),
		creator:      repo.NewCreator(aspectsTable, aspectColumns),
		deleter:      repo.NewDeleter(aspectsTable),

		conv: conv,
	}
}

// AspectCollection is an array of Entities
type AspectCollection []Entity

// Len returns the length of the collection
func (r AspectCollection) Len() int {
	return len(r)
}

// GetByID retrieves the Aspect by given ID.
func (r *pgRepository) GetByID(ctx context.Context, tenantId string, id string) (*model.Aspect, error) {
	var aspectEntity Entity
	err := r.singleGetter.Get(ctx, resource.Aspect, tenantId, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &aspectEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Aspect")
	}

	aspectModel := r.conv.FromEntity(&aspectEntity)

	return aspectModel, nil
}

// ListByResourceID lists all Aspects for an Integration Dependency with given ID.
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantId string, integrationDependencyId string) ([]*model.Aspect, error) {
	aspectCollection := AspectCollection{}

	condition := repo.NewEqualCondition("integration_dependency_id", integrationDependencyId)
	err := r.lister.ListWithSelectForUpdate(ctx, resource.Aspect, tenantId, &aspectCollection, condition)
	if err != nil {
		return nil, err
	}

	aspects := make([]*model.Aspect, 0, aspectCollection.Len())
	for _, aspect := range aspectCollection {
		aspectModel := r.conv.FromEntity(&aspect)
		aspects = append(aspects, aspectModel)
	}

	return aspects, nil
}

// Create creates an Aspect for Integration Dependency.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.Aspect) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	err := r.creator.Create(ctx, resource.Aspect, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// Delete deletes an Aspect by its ID.
func (r *pgRepository) Delete(ctx context.Context, tenantId string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.Aspect, tenantId, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteByIntegrationDependencyID deletes Aspects for an Integration Dependency with given ID.
func (r *pgRepository) DeleteByIntegrationDependencyID(ctx context.Context, tenant string, integrationDependencyId string) error {
	return r.deleter.DeleteMany(ctx, resource.Aspect, tenant, repo.Conditions{repo.NewEqualCondition("integration_dependency_id", integrationDependencyId)})
}
