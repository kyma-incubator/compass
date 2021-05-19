package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

const apiDefTable string = `"public"."api_definitions"`

var (
	tenantColumn  = "tenant_id"
	bundleColumn  = "bundle_id"
	idColumn      = "id"
	apiDefColumns = []string{"id", "tenant_id", "app_id", "package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "api_protocol", "tags", "countries", "links", "api_resource_links", "release_status",
		"sunset_date", "successor", "changelog_entries", "labels", "visibility", "disabled", "part_of_products", "line_of_business",
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "api_protocol", "tags", "countries", "links", "api_resource_links", "release_status",
		"sunset_date", "successor", "changelog_entries", "labels", "visibility", "disabled", "part_of_products", "line_of_business",
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible"}
)

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

func NewRepository(conv APIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:    repo.NewSingleGetter(resource.API, apiDefTable, tenantColumn, apiDefColumns),
		pageableQuerier: repo.NewPageableQuerier(resource.API, apiDefTable, tenantColumn, apiDefColumns),
		queryBuilder:    repo.NewQueryBuilder(resource.BundleReference, bundlereferences.BundleReferenceTable, tenantColumn, []string{bundlereferences.APIDefIDColumn}),
		lister:          repo.NewLister(resource.API, apiDefTable, tenantColumn, apiDefColumns),
		creator:         repo.NewCreator(resource.API, apiDefTable, apiDefColumns),
		updater:         repo.NewUpdater(resource.API, apiDefTable, updatableColumns, tenantColumn, idColumns),
		deleter:         repo.NewDeleter(resource.API, apiDefTable, tenantColumn),
		existQuerier:    repo.NewExistQuerier(resource.API, apiDefTable, tenantColumn),
		conv:            conv,
	}
}

type APIDefCollection []Entity

func (r APIDefCollection) Len() int {
	return len(r)
}

func (r *pgRepository) ListForBundle(ctx context.Context, tenantID string, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error) {
	return r.list(ctx, tenantID, idColumn, bundleID, pageSize, cursor)
}

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

func (r *pgRepository) list(ctx context.Context, tenant, idColumn, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error) {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition(bundleColumn, bundleID),
		repo.NewNotNullCondition(bundlereferences.APIDefIDColumn),
	}
	subquery, args, err := r.queryBuilder.BuildQuery(tenant, false, subqueryConditions...)
	if err != nil {
		return nil, err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, subquery, args),
	}

	var apiDefCollection APIDefCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, idColumn, &apiDefCollection, inOperatorConditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.APIDefinition

	for _, apiDefEnt := range apiDefCollection {
		m := r.conv.FromEntity(apiDefEnt)

		items = append(items, &m)
	}

	return &model.APIDefinitionPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.APIDefinition, error) {
	var apiDefEntity Entity
	err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	apiDefModel := r.conv.FromEntity(apiDefEntity)

	return &apiDefModel, nil
}

// the bundleID remains for backwards compatibility above in the layers; we are sure that the correct API will be fetched because there can't be two records with the same ID
func (r *pgRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error) {
	return r.GetByID(ctx, tenant, id)
}

func (r *pgRepository) Create(ctx context.Context, item *model.APIDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)
	err := r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, items []*model.APIDefinition) error {
	for index, item := range items {
		entity := r.conv.ToEntity(*item)

		err := r.creator.Create(ctx, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

func (r *pgRepository) Update(ctx context.Context, item *model.APIDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)

	return r.updater.UpdateSingle(ctx, entity)
}

func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

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
