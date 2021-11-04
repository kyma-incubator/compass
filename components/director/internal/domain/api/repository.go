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
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible", "successors", "resource_hash"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "api_protocol", "tags", "countries", "links", "api_resource_links", "release_status",
		"sunset_date", "changelog_entries", "labels", "visibility", "disabled", "part_of_products", "line_of_business",
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible", "successors", "resource_hash"}
)

// APIDefinitionConverter missing godoc
//go:generate mockery --name=APIDefinitionConverter --output=automock --outpkg=automock --case=underscore
type APIDefinitionConverter interface {
	FromEntity(entity Entity) model.APIDefinition
	ToEntity(apiModel model.APIDefinition) *Entity
}

type pgRepository struct {
	creator         repo.Creator
	singleGetter    repo.SingleGetter
	pageableQuerier repo.PageableQuerier
	queryBuilder    repo.QueryBuilder
	lister          repo.Lister
	updater         repo.Updater
	deleter         repo.Deleter
	existQuerier    repo.ExistQuerier
	conv            APIDefinitionConverter
}

// NewRepository missing godoc
func NewRepository(conv APIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:    repo.NewSingleGetter(resource.API, apiDefTable, apiDefColumns),
		pageableQuerier: repo.NewPageableQuerier(resource.API, apiDefTable, apiDefColumns),
		queryBuilder:    repo.NewQueryBuilder(resource.BundleReference, bundlereferences.BundleReferenceTable, []string{bundlereferences.APIDefIDColumn}),
		lister:          repo.NewLister(resource.API, apiDefTable, apiDefColumns),
		creator:         repo.NewCreator(resource.API, apiDefTable, apiDefColumns),
		updater:         repo.NewUpdater(resource.API, apiDefTable, updatableColumns, idColumns),
		deleter:         repo.NewDeleter(resource.API, apiDefTable),
		existQuerier:    repo.NewExistQuerier(resource.API, apiDefTable),
		conv:            conv,
	}
}

// APIDefCollection missing godoc
type APIDefCollection []Entity

// Len missing godoc
func (r APIDefCollection) Len() int {
	return len(r)
}

// ListByBundleIDs missing godoc
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
	err := r.lister.List(ctx, tenantID, &apiDefCollection, conditions...)
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

// ListByApplicationID missing godoc
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.APIDefinition, error) {
	apiCollection := APIDefCollection{}
	if err := r.lister.List(ctx, tenantID, &apiCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	apis := make([]*model.APIDefinition, 0, apiCollection.Len())
	for _, api := range apiCollection {
		apiModel := r.conv.FromEntity(api)
		apis = append(apis, &apiModel)
	}
	return apis, nil
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.APIDefinition, error) {
	var apiDefEntity Entity
	err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	apiDefModel := r.conv.FromEntity(apiDefEntity)

	return &apiDefModel, nil
}

// GetForBundle missing godoc
// the bundleID remains for backwards compatibility above in the layers; we are sure that the correct API will be fetched because there can't be two records with the same ID
func (r *pgRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error) {
	return r.GetByID(ctx, tenant, id)
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.APIDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)
	err := r.creator.Create(ctx, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateMany missing godoc
func (r *pgRepository) CreateMany(ctx context.Context, tenant string, items []*model.APIDefinition) error {
	for index, item := range items {
		entity := r.conv.ToEntity(*item)

		err := r.creator.Create(ctx, tenant, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.APIDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)

	return r.updater.UpdateSingle(ctx, tenant, entity)
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteAllByBundleID missing godoc
func (r *pgRepository) DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition(bundleColumn, bundleID),
		repo.NewNotNullCondition(bundlereferences.APIDefIDColumn),
	}
	subquery, args, err := r.queryBuilder.BuildQuery(tenantID, false, subqueryConditions...)
	if err != nil {
		return err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, subquery, args),
	}

	return r.deleter.DeleteMany(ctx, tenantID, inOperatorConditions)
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
		m := r.conv.FromEntity(apiDefEnt)
		apiDefsByAPIDefID[apiDefEnt.ID] = &m
	}

	return refsByBundleID, apiDefsByAPIDefID
}
