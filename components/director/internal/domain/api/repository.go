package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

const apiDefTable string = `"public"."api_definitions"`

var (
	bundleColumn  = "bundle_id"
	idColumn      = "id"
	apiDefColumns = []string{"id", "app_id", "package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "api_protocol", "tags", "countries", "links", "api_resource_links", "release_status",
		"sunset_date", "changelog_entries", "labels", "visibility", "disabled", "part_of_products", "line_of_business",
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible", "successors", "resource_hash", "documentation_labels"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "api_protocol", "tags", "countries", "links", "api_resource_links", "release_status",
		"sunset_date", "changelog_entries", "labels", "visibility", "disabled", "part_of_products", "line_of_business",
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible", "successors", "resource_hash", "documentation_labels"}
)

// APIDefinitionConverter converts APIDefinitions between the model.APIDefinition service-layer representation and the repo-layer representation Entity.
//go:generate mockery --name=APIDefinitionConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIDefinitionConverter interface {
	FromEntity(entity *Entity) *model.APIDefinition
	ToEntity(apiModel *model.APIDefinition) *Entity
}

type pgRepository struct {
	creator               repo.Creator
	singleGetter          repo.SingleGetter
	pageableQuerier       repo.PageableQuerier
	bundleRefQueryBuilder repo.QueryBuilderGlobal
	lister                repo.Lister
	updater               repo.Updater
	deleter               repo.Deleter
	existQuerier          repo.ExistQuerier
	conv                  APIDefinitionConverter
}

// NewRepository returns a new entity responsible for repo-layer APIDefinitions operations.
func NewRepository(conv APIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:          repo.NewSingleGetter(apiDefTable, apiDefColumns),
		pageableQuerier:       repo.NewPageableQuerier(apiDefTable, apiDefColumns),
		bundleRefQueryBuilder: repo.NewQueryBuilderGlobal(resource.BundleReference, bundlereferences.BundleReferenceTable, []string{bundlereferences.APIDefIDColumn}),
		lister:                repo.NewLister(apiDefTable, apiDefColumns),
		creator:               repo.NewCreator(apiDefTable, apiDefColumns),
		updater:               repo.NewUpdater(apiDefTable, updatableColumns, idColumns),
		deleter:               repo.NewDeleter(apiDefTable),
		existQuerier:          repo.NewExistQuerier(apiDefTable),
		conv:                  conv,
	}
}

// APIDefCollection is an array of Entities
type APIDefCollection []Entity

// Len returns the length of the collection
func (r APIDefCollection) Len() int {
	return len(r)
}

// ListByBundleIDs retrieves all APIDefinitions for a Bundle in pages. Each Bundle is extracted from the input array of bundleIDs. The input bundleReferences array is used for getting the appropriate APIDefinition IDs.
func (r *pgRepository) ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, totalCounts map[string]int, pageSize int, cursor string) ([]*model.APIDefinitionPage, error) {
	apiDefIDs := make([]string, 0, len(bundleRefs))
	for _, ref := range bundleRefs {
		apiDefIDs = append(apiDefIDs, *ref.ObjectID)
	}

	var conditions repo.Conditions
	if len(apiDefIDs) > 0 {
		conditions = repo.Conditions{
			repo.NewInConditionForStringValues("id", apiDefIDs),
		}
	}

	var apiDefCollection APIDefCollection
	err := r.lister.List(ctx, resource.API, tenantID, &apiDefCollection, conditions...)
	if err != nil {
		return nil, err
	}

	refsByBundleID, apiDefsByAPIDefID := r.groupEntitiesByID(bundleRefs, apiDefCollection)

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	apiDefPages := make([]*model.APIDefinitionPage, 0, len(bundleIDs))
	for _, bundleID := range bundleIDs {
		ids := getAPIDefIDsForBundle(refsByBundleID[bundleID])
		apiDefs := getAPIDefsForBundle(ids, apiDefsByAPIDefID)

		hasNextPage := false
		endCursor := ""
		if totalCounts[bundleID] > offset+len(apiDefs) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		apiDefPages = append(apiDefPages, &model.APIDefinitionPage{Data: apiDefs, TotalCount: totalCounts[bundleID], PageInfo: page})
	}

	return apiDefPages, nil
}

// ListByApplicationID lists all APIDefinitions for a given application ID.
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.APIDefinition, error) {
	apiCollection := APIDefCollection{}
	if err := r.lister.ListWithSelectForUpdate(ctx, resource.API, tenantID, &apiCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	apis := make([]*model.APIDefinition, 0, apiCollection.Len())
	for _, api := range apiCollection {
		apiModel := r.conv.FromEntity(&api)
		apis = append(apis, apiModel)
	}
	return apis, nil
}

// GetByID retrieves the APIDefinition with matching ID from the Compass storage.
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.APIDefinition, error) {
	var apiDefEntity Entity
	err := r.singleGetter.Get(ctx, resource.API, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	apiDefModel := r.conv.FromEntity(&apiDefEntity)

	return apiDefModel, nil
}

// GetForBundle gets an APIDefinition by its id.
// the bundleID remains for backwards compatibility above in the layers; we are sure that the correct API will be fetched because there can't be two records with the same ID
func (r *pgRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error) {
	return r.GetByID(ctx, tenant, id)
}

// Create creates an APIDefinition.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.APIDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creator.Create(ctx, resource.API, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateMany creates many APIDefinitions.
func (r *pgRepository) CreateMany(ctx context.Context, tenant string, items []*model.APIDefinition) error {
	for index, item := range items {
		entity := r.conv.ToEntity(item)

		err := r.creator.Create(ctx, resource.API, tenant, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

// Update updates an APIDefinition.
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.APIDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, resource.API, tenant, entity)
}

// Exists checks if an APIDefinition with a given ID exists.
func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.API, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Delete deletes an APIDefinition by its ID.
func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.API, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteAllByBundleID deletes all APIDefinitions for a given bundle ID.
func (r *pgRepository) DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition(bundleColumn, bundleID),
		repo.NewNotNullCondition(bundlereferences.APIDefIDColumn),
	}
	subquery, args, err := r.bundleRefQueryBuilder.BuildQueryGlobal(false, subqueryConditions...)
	if err != nil {
		return err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, subquery, args),
	}

	return r.deleter.DeleteMany(ctx, resource.API, tenantID, inOperatorConditions)
}

func getAPIDefIDsForBundle(refs []*model.BundleReference) []string {
	result := make([]string, 0, len(refs))
	for _, ref := range refs {
		result = append(result, *ref.ObjectID)
	}
	return result
}

func getAPIDefsForBundle(ids []string, defs map[string]*model.APIDefinition) []*model.APIDefinition {
	result := make([]*model.APIDefinition, 0, len(ids))
	if len(defs) > 0 {
		for _, id := range ids {
			result = append(result, defs[id])
		}
	}
	return result
}

func (r *pgRepository) groupEntitiesByID(bundleRefs []*model.BundleReference, apiDefCollection APIDefCollection) (map[string][]*model.BundleReference, map[string]*model.APIDefinition) {
	refsByBundleID := map[string][]*model.BundleReference{}
	for _, ref := range bundleRefs {
		refsByBundleID[*ref.BundleID] = append(refsByBundleID[*ref.BundleID], ref)
	}

	apiDefsByAPIDefID := map[string]*model.APIDefinition{}
	for _, apiDefEnt := range apiDefCollection {
		m := r.conv.FromEntity(&apiDefEnt)
		apiDefsByAPIDefID[apiDefEnt.ID] = m
	}

	return refsByBundleID, apiDefsByAPIDefID
}
