package capability

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const capabilityTable string = `"public"."capabilities"`

var (
	capabilityColumns = []string{"id", "app_id", "app_template_version_id", "package_id", "name", "description", "ord_id", "local_tenant_id",
		"short_description", "system_instance_aware", "tags", "links", "release_status", "labels", "visibility",
		"version_value", "ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash", "documentation_labels", "correlation_ids"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"package_id", "name", "description", "ord_id", "local_tenant_id",
		"short_description", "system_instance_aware", "tags", "links", "release_status",
		"labels", "visibility", "version_value", "ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash", "documentation_labels", "correlation_ids"}
)

// CapabilityConverter converts Capabilities between the model.Capability service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=CapabilityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type CapabilityConverter interface {
	FromEntity(entity *Entity) *model.Capability
	ToEntity(apiModel *model.Capability) *Entity
}

type pgRepository struct {
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	creator            repo.Creator
	creatorGlobal      repo.CreatorGlobal
	updater            repo.Updater
	updaterGlobal      repo.UpdaterGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal
	conv               CapabilityConverter
}

// NewRepository returns a new entity responsible for repo-layer Capabilities operations.
func NewRepository(conv CapabilityConverter) *pgRepository {
	return &pgRepository{
		lister:             repo.NewLister(capabilityTable, capabilityColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Capability, capabilityTable, capabilityColumns),
		singleGetter:       repo.NewSingleGetter(capabilityTable, capabilityColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Capability, capabilityTable, capabilityColumns),
		creator:            repo.NewCreator(capabilityTable, capabilityColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.Capability, capabilityTable, capabilityColumns),
		updater:            repo.NewUpdater(capabilityTable, capabilityColumns, idColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Capability, capabilityTable, updatableColumns, idColumns),
		deleter:            repo.NewDeleter(capabilityTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Capability, capabilityTable),
		conv:               conv,
	}
}

// CapabilityCollection is an array of Entities
type CapabilityCollection []Entity

// Len returns the length of the collection
func (r CapabilityCollection) Len() int {
	return len(r)
}

// ListByResourceID lists all Capabilities for a given resource ID and resource type.
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.Capability, error) {
	capabilityCollection := CapabilityCollection{}

	var condition repo.Condition
	var err error
	if resourceType == resource.Application {
		condition = repo.NewEqualCondition("app_id", resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.Capability, tenantID, &capabilityCollection, condition)
	} else {
		condition = repo.NewEqualCondition("app_template_version_id", resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &capabilityCollection, condition)
	}
	if err != nil {
		return nil, err
	}

	capabilities := make([]*model.Capability, 0, capabilityCollection.Len())
	for _, capability := range capabilityCollection {
		capabilityModel := r.conv.FromEntity(&capability)
		capabilities = append(capabilities, capabilityModel)
	}

	return capabilities, nil
}

// GetByID retrieves a Capability by ID.
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.Capability, error) {
	var capabilityEntity Entity
	err := r.singleGetter.Get(ctx, resource.API, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &capabilityEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Capability")
	}

	capabilityModel := r.conv.FromEntity(&capabilityEntity)

	return capabilityModel, nil
}

// GetByIDGlobal retrieves a Capability by ID without tenant isolation
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.Capability, error) {
	var capabilityEntity Entity
	err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &capabilityEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	capabilityModel := r.conv.FromEntity(&capabilityEntity)

	return capabilityModel, nil
}

// Create creates a Capability.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.Capability) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creator.Create(ctx, resource.Capability, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateGlobal creates a Capability.
func (r *pgRepository) CreateGlobal(ctx context.Context, item *model.Capability) error {
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

// Update updates an Capability.
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.Capability) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, resource.Capability, tenant, entity)
}

// UpdateGlobal updates an existing Capability without tenant isolation.
func (r *pgRepository) UpdateGlobal(ctx context.Context, item *model.Capability) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Delete deletes a Capability by its ID.
func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.Capability, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal deletes a Capability by its ID without tenant isolation.
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}
