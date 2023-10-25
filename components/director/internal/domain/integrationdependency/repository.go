package integrationdependency

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const integrationDependencyTable string = `"public"."integration_dependencies"`

var (
	idColumns                    = []string{"id"}
	integrationDependencyColumns = []string{"id", "app_id", "app_template_version_id", "ord_id", "local_tenant_id", "correlation_ids", "name", "short_description", "description", "package_id", "last_update", "visibility",
		"release_status", "sunset_date", "successors", "mandatory", "related_integration_dependencies", "links", "tags", "labels", "documentation_labels", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal",
		"ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash"}
	updatableColumns = []string{"ord_id", "local_tenant_id", "correlation_ids", "name", "short_description", "description", "package_id", "last_update", "visibility",
		"release_status", "sunset_date", "successors", "mandatory", "related_integration_dependencies", "links", "tags", "labels", "documentation_labels", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal",
		"ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash"}
)

// IntegrationDependencyConverter converts IntegrationDependencies between the model.IntegrationDependency service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=IntegrationDependencyConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDependencyConverter interface {
	FromEntity(entity *Entity) *model.IntegrationDependency
	ToEntity(integrationDependencyModel *model.IntegrationDependency) *Entity
}

type pgRepository struct {
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	creator            repo.Creator
	creatorGlobal      repo.CreatorGlobal
	updater            repo.Updater
	updaterGlobal      repo.UpdaterGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal

	conv IntegrationDependencyConverter
}

// NewRepository returns a new entity responsible for repo-layer IntegrationDependencies operations.
func NewRepository(conv IntegrationDependencyConverter) *pgRepository {
	return &pgRepository{
		singleGetter:       repo.NewSingleGetter(integrationDependencyTable, integrationDependencyColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.IntegrationDependency, integrationDependencyTable, integrationDependencyColumns),
		lister:             repo.NewLister(integrationDependencyTable, integrationDependencyColumns),
		listerGlobal:       repo.NewListerGlobal(resource.IntegrationDependency, integrationDependencyTable, integrationDependencyColumns),
		creator:            repo.NewCreator(integrationDependencyTable, integrationDependencyColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.IntegrationDependency, integrationDependencyTable, integrationDependencyColumns),
		updater:            repo.NewUpdater(integrationDependencyTable, updatableColumns, idColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.IntegrationDependency, integrationDependencyTable, updatableColumns, idColumns),
		deleter:            repo.NewDeleter(integrationDependencyTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.IntegrationDependency, integrationDependencyTable),

		conv: conv,
	}
}

// IntegrationDependencyCollection is an array of Entities
type IntegrationDependencyCollection []Entity

// Len returns the length of the collection
func (r IntegrationDependencyCollection) Len() int {
	return len(r)
}

// GetByID retrieves the IntegrationDependency with matching ID from the Compass storage.
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.IntegrationDependency, error) {
	var integrationDependencyEntity Entity
	err := r.singleGetter.Get(ctx, resource.IntegrationDependency, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &integrationDependencyEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Integration Dependency")
	}

	integrationDependencyModel := r.conv.FromEntity(&integrationDependencyEntity)

	return integrationDependencyModel, nil
}

// GetByIDGlobal gets an IntegrationDependency by ID without tenant isolation
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.IntegrationDependency, error) {
	var integrationDependencyEntity Entity
	err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &integrationDependencyEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Integration Dependency")
	}

	integrationDependencyModel := r.conv.FromEntity(&integrationDependencyEntity)

	return integrationDependencyModel, nil
}

// ListByResourceID lists all IntegrationDependencies for a given resource ID and resource type.
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.IntegrationDependency, error) {
	integrationDependencyCollection := IntegrationDependencyCollection{}

	var condition repo.Condition
	var err error
	if resourceType == resource.Application {
		condition = repo.NewEqualCondition("app_id", resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.IntegrationDependency, tenantID, &integrationDependencyCollection, condition)
	} else {
		condition = repo.NewEqualCondition("app_template_version_id", resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &integrationDependencyCollection, condition)
	}
	if err != nil {
		return nil, err
	}

	integrationDependencies := make([]*model.IntegrationDependency, 0, integrationDependencyCollection.Len())
	for _, integrationDependency := range integrationDependencyCollection {
		integrationDependencyModel := r.conv.FromEntity(&integrationDependency)
		integrationDependencies = append(integrationDependencies, integrationDependencyModel)
	}

	return integrationDependencies, nil
}

// Create creates an IntegrationDependency.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.IntegrationDependency) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creator.Create(ctx, resource.IntegrationDependency, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateGlobal creates an IntegrationDependency without tenant isolation.
func (r *pgRepository) CreateGlobal(ctx context.Context, item *model.IntegrationDependency) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creatorGlobal.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// Update updates an IntegrationDependency.
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.IntegrationDependency) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, resource.IntegrationDependency, tenant, entity)
}

// UpdateGlobal updates an existing IntegrationDependency without tenant isolation.
func (r *pgRepository) UpdateGlobal(ctx context.Context, item *model.IntegrationDependency) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Delete deletes an IntegrationDependency by its ID.
func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.IntegrationDependency, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal deletes an IntegrationDependency by its ID without tenant isolation.
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}
