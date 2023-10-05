package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const eventAPIDefTable string = `"public"."event_api_definitions"`

var (
	idColumn                 = "id"
	appColumn                = "app_id"
	appTemplateVersionColumn = "app_template_version_id"
	bundleColumn             = "bundle_id"
	eventDefColumns          = []string{idColumn, appColumn, "app_template_version_id", "package_id", "name", "description", "group_name", "ord_id", "local_tenant_id",
		"short_description", "system_instance_aware", "policy_level", "custom_policy_level", "changelog_entries", "links", "tags", "countries", "release_status",
		"sunset_date", "labels", "visibility", "disabled", "part_of_products", "line_of_business", "industry", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "extensible", "successors", "resource_hash", "documentation_labels", "correlation_ids"}
	idColumns        = []string{idColumn}
	updatableColumns = []string{"package_id", "name", "description", "group_name", "ord_id", "local_tenant_id",
		"short_description", "system_instance_aware", "policy_level", "custom_policy_level", "changelog_entries", "links", "tags", "countries", "release_status",
		"sunset_date", "labels", "visibility", "disabled", "part_of_products", "line_of_business", "industry", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "extensible", "successors", "resource_hash", "documentation_labels", "correlation_ids"}
)

// EventAPIDefinitionConverter converts EventDefinitions between the model.EventDefinition service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=EventAPIDefinitionConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventAPIDefinitionConverter interface {
	FromEntity(entity *Entity) *model.EventDefinition
	ToEntity(apiModel *model.EventDefinition) *Entity
}

type pgRepository struct {
	singleGetter          repo.SingleGetter
	singleGetterGlobal    repo.SingleGetterGlobal
	bundleRefQueryBuilder repo.QueryBuilderGlobal
	lister                repo.Lister
	listerGlobal          repo.ListerGlobal
	creator               repo.Creator
	creatorGlobal         repo.CreatorGlobal
	updater               repo.Updater
	updaterGlobal         repo.UpdaterGlobal
	deleter               repo.Deleter
	deleterGlobal         repo.DeleterGlobal
	existQuerier          repo.ExistQuerier
	pageableQuerier       repo.PageableQuerier
	conv                  EventAPIDefinitionConverter
}

// NewRepository returns a new entity responsible for repo-layer EventDefinitions operations.
func NewRepository(conv EventAPIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:          repo.NewSingleGetter(eventAPIDefTable, eventDefColumns),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(resource.EventDefinition, eventAPIDefTable, eventDefColumns),
		bundleRefQueryBuilder: repo.NewQueryBuilderGlobal(resource.BundleReference, bundlereferences.BundleReferenceTable, []string{bundlereferences.EventDefIDColumn}),
		lister:                repo.NewLister(eventAPIDefTable, eventDefColumns),
		listerGlobal:          repo.NewListerGlobal(resource.EventDefinition, eventAPIDefTable, eventDefColumns),
		creator:               repo.NewCreator(eventAPIDefTable, eventDefColumns),
		creatorGlobal:         repo.NewCreatorGlobal(resource.EventDefinition, eventAPIDefTable, eventDefColumns),
		updater:               repo.NewUpdater(eventAPIDefTable, updatableColumns, idColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(resource.EventDefinition, eventAPIDefTable, updatableColumns, idColumns),
		deleter:               repo.NewDeleter(eventAPIDefTable),
		deleterGlobal:         repo.NewDeleterGlobal(resource.EventDefinition, eventAPIDefTable),
		existQuerier:          repo.NewExistQuerier(eventAPIDefTable),
		pageableQuerier:       repo.NewPageableQuerier(eventAPIDefTable, eventDefColumns),
		conv:                  conv,
	}
}

// EventAPIDefCollection is an array of Entities
type EventAPIDefCollection []Entity

// Len returns the length of the collection
func (r EventAPIDefCollection) Len() int {
	return len(r)
}

// GetByID retrieves the EventDefinition with matching ID from the Compass storage.
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.EventDefinition, error) {
	var eventAPIDefEntity Entity
	err := r.singleGetter.Get(ctx, resource.EventDefinition, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EventDefinition with id %s", id)
	}

	eventAPIDefModel := r.conv.FromEntity(&eventAPIDefEntity)
	return eventAPIDefModel, nil
}

// GetByIDGlobal retrieves the EventDefinition with matching ID from the Compass storage.
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.EventDefinition, error) {
	var eventAPIDefEntity Entity
	err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EventDefinition with id %s", id)
	}

	eventAPIDefModel := r.conv.FromEntity(&eventAPIDefEntity)
	return eventAPIDefModel, nil
}

// GetForBundle gets an EventDefinition by its id.
// the bundleID remains for backwards compatibility above in the layers; we are sure that the correct Event will be fetched because there can't be two records with the same ID
func (r *pgRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.EventDefinition, error) {
	return r.GetByID(ctx, tenant, id)
}

// GetByApplicationID retrieves the EventDefinition with matching ID and Application ID from the Compass storage.
func (r *pgRepository) GetByApplicationID(ctx context.Context, tenantID string, id, appID string) (*model.EventDefinition, error) {
	var eventDefEntity Entity
	err := r.singleGetter.Get(ctx, resource.EventDefinition, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id), repo.NewEqualCondition(appColumn, appID)}, repo.NoOrderBy, &eventDefEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EventDefinition for Application ID %s", appID)
	}

	eventDefModel := r.conv.FromEntity(&eventDefEntity)

	return eventDefModel, nil
}

// ListByApplicationIDPage lists all EventDefinitions for a given application ID with paging.
func (r *pgRepository) ListByApplicationIDPage(ctx context.Context, tenantID string, appID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	var apiDefCollection EventAPIDefCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, resource.EventDefinition, tenantID, pageSize, cursor, idColumn, &apiDefCollection, repo.NewEqualCondition("app_id", appID))

	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	items := make([]*model.EventDefinition, 0, len(apiDefCollection))
	for _, api := range apiDefCollection {
		m := r.conv.FromEntity(&api)
		items = append(items, m)
	}

	return &model.EventDefinitionPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListByBundleIDs retrieves all EventDefinitions for a Bundle in pages. Each Bundle is extracted from the input array of bundleIDs. The input bundleReferences array is used for getting the appropriate EventDefinition IDs.
func (r *pgRepository) ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, totalCounts map[string]int, pageSize int, cursor string) ([]*model.EventDefinitionPage, error) {
	eventDefIDs := make([]string, 0, len(bundleRefs))
	for _, ref := range bundleRefs {
		eventDefIDs = append(eventDefIDs, *ref.ObjectID)
	}

	var conditions repo.Conditions
	if len(eventDefIDs) > 0 {
		conditions = repo.Conditions{
			repo.NewInConditionForStringValues(idColumn, eventDefIDs),
		}
	}

	var eventCollection EventAPIDefCollection
	err := r.lister.List(ctx, resource.EventDefinition, tenantID, &eventCollection, conditions...)
	if err != nil {
		return nil, err
	}

	refsByBundleID, eventDefsByEventDefID := r.groupEntitiesByID(bundleRefs, eventCollection)

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	eventDefPages := make([]*model.EventDefinitionPage, 0, len(bundleIDs))
	for _, bundleID := range bundleIDs {
		ids := getEventDefIDsForBundle(refsByBundleID[bundleID])
		eventDefs := getEventDefsForBundle(ids, eventDefsByEventDefID)
		hasNextPage := false
		endCursor := ""
		if totalCounts[bundleID] > offset+len(eventDefs) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		eventDefPages = append(eventDefPages, &model.EventDefinitionPage{Data: eventDefs, TotalCount: totalCounts[bundleID], PageInfo: page})
	}

	return eventDefPages, nil
}

// ListByResourceID lists all EventDefinitions for a given resource ID and resource.Type
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.EventDefinition, error) {
	eventCollection := EventAPIDefCollection{}

	var condition repo.Condition
	var err error
	if resourceType == resource.Application {
		condition = repo.NewEqualCondition(appColumn, resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.EventDefinition, tenantID, &eventCollection, condition)
	} else {
		condition = repo.NewEqualCondition(appTemplateVersionColumn, resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &eventCollection, condition)
	}
	if err != nil {
		return nil, err
	}

	events := make([]*model.EventDefinition, 0, eventCollection.Len())
	for _, event := range eventCollection {
		eventModel := r.conv.FromEntity(&event)
		events = append(events, eventModel)
	}
	return events, nil
}

// Create creates an EventDefinition.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.EventDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	log.C(ctx).Debugf("Persisting Event-Definition entity with id %s to db", item.ID)
	err := r.creator.Create(ctx, resource.EventDefinition, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateGlobal creates an EventDefinition without tenant isolation.
func (r *pgRepository) CreateGlobal(ctx context.Context, item *model.EventDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	log.C(ctx).Debugf("Persisting Event-Definition entity with id %s to db", item.ID)
	err := r.creatorGlobal.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateMany creates many EventDefinitions.
func (r *pgRepository) CreateMany(ctx context.Context, tenant string, items []*model.EventDefinition) error {
	for index, item := range items {
		entity := r.conv.ToEntity(item)
		err := r.creator.Create(ctx, resource.EventDefinition, tenant, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

// Update updates an EventDefinition.
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.EventDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, resource.EventDefinition, tenant, entity)
}

func (r *pgRepository) UpdateGlobal(ctx context.Context, item *model.EventDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Exists checks if an EventDefinition with a given ID exists.
func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.EventDefinition, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// Delete deletes an EventDefinition by its ID.
func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.EventDefinition, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// DeleteGlobal deletes an EventDefinition by its ID without tenant isolation.
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// DeleteAllByBundleID deletes all EventDefinitions for a given bundle ID.
func (r *pgRepository) DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition(bundleColumn, bundleID),
		repo.NewNotNullCondition(bundlereferences.EventDefIDColumn),
	}
	subquery, args, err := r.bundleRefQueryBuilder.BuildQueryGlobal(false, subqueryConditions...)
	if err != nil {
		return err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, subquery, args),
	}

	return r.deleter.DeleteMany(ctx, resource.EventDefinition, tenantID, inOperatorConditions)
}

func getEventDefsForBundle(ids []string, defs map[string]*model.EventDefinition) []*model.EventDefinition {
	result := make([]*model.EventDefinition, 0, len(ids))
	if len(defs) > 0 {
		for _, id := range ids {
			result = append(result, defs[id])
		}
	}
	return result
}

func getEventDefIDsForBundle(refs []*model.BundleReference) []string {
	result := make([]string, 0, len(refs))
	for _, ref := range refs {
		result = append(result, *ref.ObjectID)
	}
	return result
}

func (r *pgRepository) groupEntitiesByID(bundleRefs []*model.BundleReference, eventCollectionCollection EventAPIDefCollection) (map[string][]*model.BundleReference, map[string]*model.EventDefinition) {
	refsByBundleID := map[string][]*model.BundleReference{}
	for _, ref := range bundleRefs {
		refsByBundleID[*ref.BundleID] = append(refsByBundleID[*ref.BundleID], ref)
	}

	eventsByAPIDefID := map[string]*model.EventDefinition{}
	for _, eventEnt := range eventCollectionCollection {
		m := r.conv.FromEntity(&eventEnt)
		eventsByAPIDefID[eventEnt.ID] = m
	}

	return refsByBundleID, eventsByAPIDefID
}
