package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const eventAPIDefTable string = `"public"."event_api_definitions"`

var (
	idColumn        = "id"
	tenantColumn    = "tenant_id"
	appColumn       = "app_id"
	bundleColumn    = "bundle_id"
	eventDefColumns = []string{idColumn, tenantColumn, appColumn, "package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "changelog_entries", "links", "tags", "countries", "release_status",
		"sunset_date", "successor", "labels", "visibility", "disabled", "part_of_products", "line_of_business", "industry", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "extensible"}
	idColumns        = []string{idColumn}
	updatableColumns = []string{"package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "changelog_entries", "links", "tags", "countries", "release_status",
		"sunset_date", "successor", "labels", "visibility", "disabled", "part_of_products", "line_of_business", "industry", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "extensible"}
)

//go:generate mockery --name=EventAPIDefinitionConverter --output=automock --outpkg=automock --case=underscore
type EventAPIDefinitionConverter interface {
	FromEntity(entity Entity) model.EventDefinition
	ToEntity(apiModel model.EventDefinition) Entity
}

type pgRepository struct {
	singleGetter    repo.SingleGetter
	pageableQuerier repo.PageableQuerier
	queryBuilder    repo.QueryBuilder
	lister          repo.Lister
	creator         repo.Creator
	updater         repo.Updater
	deleter         repo.Deleter
	existQuerier    repo.ExistQuerier
	conv            EventAPIDefinitionConverter
}

func NewRepository(conv EventAPIDefinitionConverter) *pgRepository {
	return &pgRepository{
		singleGetter:    repo.NewSingleGetter(resource.EventDefinition, eventAPIDefTable, tenantColumn, eventDefColumns),
		pageableQuerier: repo.NewPageableQuerier(resource.EventDefinition, eventAPIDefTable, tenantColumn, eventDefColumns),
		queryBuilder:    repo.NewQueryBuilder(resource.BundleReference, bundlereferences.BundleReferenceTable, tenantColumn, []string{bundlereferences.EventDefIDColumn}),
		lister:          repo.NewLister(resource.EventDefinition, eventAPIDefTable, tenantColumn, eventDefColumns),
		creator:         repo.NewCreator(resource.EventDefinition, eventAPIDefTable, eventDefColumns),
		updater:         repo.NewUpdater(resource.EventDefinition, eventAPIDefTable, updatableColumns, tenantColumn, idColumns),
		deleter:         repo.NewDeleter(resource.EventDefinition, eventAPIDefTable, tenantColumn),
		existQuerier:    repo.NewExistQuerier(resource.EventDefinition, eventAPIDefTable, tenantColumn),
		conv:            conv,
	}
}

type EventAPIDefCollection []Entity

func (r EventAPIDefCollection) Len() int {
	return len(r)
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.EventDefinition, error) {
	var eventAPIDefEntity Entity
	err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &eventAPIDefEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EventDefinition with id %s", id)
	}

	eventAPIDefModel := r.conv.FromEntity(eventAPIDefEntity)
	return &eventAPIDefModel, nil
}

// the bundleID remains for backwards compatibility above in the layers; we are sure that the correct Event will be fetched because there can't be two records with the same ID
func (r *pgRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.EventDefinition, error) {
	return r.GetByID(ctx, tenant, id)
}

func (r *pgRepository) ListForBundle(ctx context.Context, tenantID string, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	return r.list(ctx, tenantID, idColumn, bundleID, pageSize, cursor)
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.EventDefinition, error) {
	eventCollection := EventAPIDefCollection{}
	if err := r.lister.List(ctx, tenantID, &eventCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	events := make([]*model.EventDefinition, 0, eventCollection.Len())
	for _, event := range eventCollection {
		eventModel := r.conv.FromEntity(event)
		events = append(events, &eventModel)
	}
	return events, nil
}

func (r *pgRepository) list(ctx context.Context, tenant, idColumn, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition(bundleColumn, bundleID),
		repo.NewNotNullCondition(bundlereferences.EventDefIDColumn),
	}
	subquery, args, err := r.queryBuilder.BuildQuery(tenant, false, subqueryConditions...)
	if err != nil {
		return nil, err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, subquery, args),
	}

	var eventCollection EventAPIDefCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, idColumn, &eventCollection, inOperatorConditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.EventDefinition

	for _, eventEnt := range eventCollection {
		m := r.conv.FromEntity(eventEnt)
		items = append(items, &m)
	}

	return &model.EventDefinitionPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.EventDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)

	log.C(ctx).Debugf("Persisting Event-Definition entity with id %s to db", item.ID)
	err := r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, items []*model.EventDefinition) error {
	for index, item := range items {
		entity := r.conv.ToEntity(*item)
		err := r.creator.Create(ctx, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

func (r *pgRepository) Update(ctx context.Context, item *model.EventDefinition) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)

	return r.updater.UpdateSingle(ctx, entity)
}

func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

func (r *pgRepository) DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition(bundleColumn, bundleID),
		repo.NewNotNullCondition(bundlereferences.EventDefIDColumn),
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
