package integrationdependency

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const integrationDependencyTable string = `public.integration_dependencies`

var (
	idColumns                    = []string{"id"}
	integrationDependencyColumns = []string{"id", "app_id", "app_template_version_id", "ord_id", "local_tenant_id", "correlation_ids", "title", "short_description", "description", "package_id", "last_update", "visibility",
		"release_status", "sunset_date", "successors", "mandatory", "related_integration_dependencies", "links", "tags", "labels", "documentation_labels", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal",
		"ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash"}
	updatableColumns = []string{"ord_id", "local_tenant_id", "correlation_ids", "title", "short_description", "description", "package_id", "last_update", "visibility",
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
	switch resourceType {
	case resource.Application:
		condition = repo.NewEqualCondition("app_id", resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.IntegrationDependency, tenantID, &integrationDependencyCollection, condition)
	case resource.ApplicationTemplateVersion:
		condition = repo.NewEqualCondition("app_template_version_id", resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &integrationDependencyCollection, condition)
	case resource.Package:
		condition = repo.NewEqualCondition("package_id", resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.IntegrationDependency, tenantID, &integrationDependencyCollection, condition)
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

// ListByApplicationIDs retrieves all Integration Dependencies for an Application in pages. Each Application is extracted from the input array of applicationIDs. The input aspects array is used for getting the appropriate Integration Dependencies IDs.
func (r *pgRepository) ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, aspects []*model.Aspect, totalCounts map[string]int, pageSize int, cursor string) ([]*model.IntegrationDependencyPage, error) {
	integrationDependenciesIDs := make([]string, 0, len(aspects))
	for _, aspect := range aspects {
		integrationDependenciesIDs = append(integrationDependenciesIDs, aspect.IntegrationDependencyID)
	}

	var conditions repo.Conditions
	if len(integrationDependenciesIDs) > 0 {
		conditions = repo.Conditions{
			repo.NewInConditionForStringValues("id", integrationDependenciesIDs),
		}
	}

	var integrationDependencyCollection IntegrationDependencyCollection
	err := r.lister.List(ctx, resource.IntegrationDependency, tenantID, &integrationDependencyCollection, conditions...)
	if err != nil {
		return nil, err
	}

	aspectsByApplicationID, integrationDependenciesByIntegrationDependencyID := r.groupEntitiesByID(aspects, integrationDependencyCollection)

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	integrationDependencyPages := make([]*model.IntegrationDependencyPage, 0, len(applicationIDs))
	for _, appID := range applicationIDs {
		ids := getIntegrationDependencyIDsForApplication(aspectsByApplicationID[appID])
		integrationDependencies := getIntegrationDependenciesForApplication(ids, integrationDependenciesByIntegrationDependencyID)

		hasNextPage := false
		endCursor := ""
		if totalCounts[appID] > offset+len(integrationDependencies) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		integrationDependencyPages = append(integrationDependencyPages, &model.IntegrationDependencyPage{Data: integrationDependencies, TotalCount: totalCounts[appID], PageInfo: page})
	}

	return integrationDependencyPages, nil
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

func getIntegrationDependencyIDsForApplication(aspects []*model.Aspect) []string {
	integrationDependencyIDExists := make(map[string]bool)
	result := make([]string, 0, len(aspects))

	for _, aspect := range aspects {
		id := aspect.IntegrationDependencyID
		if !integrationDependencyIDExists[id] {
			result = append(result, id)
			integrationDependencyIDExists[id] = true
		}
	}
	return result
}

func getIntegrationDependenciesForApplication(ids []string, integrationDependencies map[string]*model.IntegrationDependency) []*model.IntegrationDependency {
	result := make([]*model.IntegrationDependency, 0, len(ids))
	if len(integrationDependencies) > 0 {
		for _, id := range ids {
			result = append(result, integrationDependencies[id])
		}
	}
	return result
}

func (r *pgRepository) groupEntitiesByID(aspects []*model.Aspect, integrationDependencyCollection IntegrationDependencyCollection) (map[string][]*model.Aspect, map[string]*model.IntegrationDependency) {
	aspectsByApplicationID := map[string][]*model.Aspect{}
	for _, aspect := range aspects {
		aspectsByApplicationID[*aspect.ApplicationID] = append(aspectsByApplicationID[*aspect.ApplicationID], aspect)
	}

	integrationDependenciesByIntegrationDependencyID := map[string]*model.IntegrationDependency{}
	for _, integrationDependencyEnt := range integrationDependencyCollection {
		m := r.conv.FromEntity(&integrationDependencyEnt)
		integrationDependenciesByIntegrationDependencyID[integrationDependencyEnt.ID] = m
	}

	return aspectsByApplicationID, integrationDependenciesByIntegrationDependencyID
}
