package bundleinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.bundle_instance_auths`
const ScenariosViewName = `public.bundle_instance_auths_with_labels`

var (
	tenantColumn     = "tenant_id"
	idColumns        = []string{"id"}
	updatableColumns = []string{"auth_value", "status_condition", "status_timestamp", "status_message", "status_reason"}
	tableColumns     = []string{"id", "tenant_id", "bundle_id", "context", "input_params", "auth_value", "status_condition", "status_timestamp", "status_message", "status_reason", "runtime_id", "runtime_context_id"}
)

//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	ToEntity(in model.BundleInstanceAuth) (Entity, error)
	FromEntity(entity Entity) (model.BundleInstanceAuth, error)
}

type scenariosView struct {
	qBuilder repo.QueryBuilder
	repo.ExistQuerier
}

type repository struct {
	creator       repo.Creator
	singleGetter  repo.SingleGetter
	lister        repo.Lister
	updater       repo.Updater
	deleter       repo.Deleter
	conv          EntityConverter
	scenariosView *scenariosView
}

func NewRepository(conv EntityConverter) *repository {
	scenarioQueryBuilder := repo.NewQueryBuilder(resource.Label, ScenariosViewName, tenantColumn, []string{"bundle_instance_auth_id"})

	return &repository{
		creator:      repo.NewCreator(resource.BundleInstanceAuth, tableName, tableColumns),
		singleGetter: repo.NewSingleGetter(resource.BundleInstanceAuth, tableName, tenantColumn, tableColumns),
		lister:       repo.NewLister(resource.BundleInstanceAuth, tableName, tenantColumn, tableColumns),
		deleter:      repo.NewDeleter(resource.BundleInstanceAuth, tableName, tenantColumn),
		updater:      repo.NewUpdater(resource.BundleInstanceAuth, tableName, updatableColumns, tenantColumn, idColumns),
		conv:         conv,
		scenariosView: &scenariosView{
			qBuilder:     scenarioQueryBuilder,
			ExistQuerier: repo.NewExistQuerier(resource.BundleInstanceAuth, ScenariosViewName, tenantColumn)},
	}
}

func (r *repository) Create(ctx context.Context, item *model.BundleInstanceAuth) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting BundleInstanceAuth model to entity")
	}

	log.C(ctx).Debugf("Persisting BundleInstanceAuth entity with id %s to db", item.ID)
	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrapf(err, "while saving entity with id %s to db", item.ID)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, tenantID string, id string) (*model.BundleInstanceAuth, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting BundleInstanceAuth entity to model")
	}

	return &itemModel, nil
}

func (r *repository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.BundleInstanceAuth, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("bundle_id", bundleID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return &bndlModel, nil
}

func (r *repository) ListByBundleID(ctx context.Context, tenantID string, bundleID string) ([]*model.BundleInstanceAuth, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("bundle_id", bundleID),
	}

	err := r.lister.List(ctx, tenantID, &entities, conditions...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

func (r *repository) ListByRuntimeID(ctx context.Context, tenantID string, runtimeID string) ([]*model.BundleInstanceAuth, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("runtime_id", runtimeID),
	}

	err := r.lister.List(ctx, tenantID, &entities, conditions...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

func (r *repository) Update(ctx context.Context, item *model.BundleInstanceAuth) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	log.C(ctx).Debugf("Updating BundleInstanceAuth entity with id %s in db", item.ID)
	return r.updater.UpdateSingle(ctx, entity)
}

func (r *repository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) multipleFromEntities(entities Collection) ([]*model.BundleInstanceAuth, error) {
	var items []*model.BundleInstanceAuth
	for _, ent := range entities {
		m, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating BundleInstanceAuth model from entity")
		}
		items = append(items, &m)
	}
	return items, nil
}

func (r *repository) GetForAppAndAnyMatchingScenarios(ctx context.Context, tenant, appId string, scenarios []string) ([]*model.BundleInstanceAuth, error) {
	return r.getForObjectAndAnyMatchingScenarios(ctx, tenant, "app_id", appId, scenarios)
}

func (r *repository) GetForRuntimeAndAnyMatchingScenarios(ctx context.Context, tenant, runtimeId string, scenarios []string) ([]*model.BundleInstanceAuth, error) {
	return r.getForObjectAndAnyMatchingScenarios(ctx, tenant, "runtime_id", runtimeId, scenarios)
}

func (r *repository) getForObjectAndAnyMatchingScenarios(ctx context.Context, tenant, objectColumn, objectId string, scenarios []string) ([]*model.BundleInstanceAuth, error) {
	values := make([]interface{}, 0, len(scenarios))
	for _, scenario := range scenarios {
		values = append(values, scenario)
	}

	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition("key", model.ScenariosKey),
		repo.NewEqualCondition(objectColumn, objectId),
		repo.NewJSONArrAnyMatchCondition("value", values),
	}

	subquery, args, err := r.scenariosView.qBuilder.BuildQuery(tenant, false, subqueryConditions...)
	if err != nil {
		return nil, err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery("id", subquery, args),
	}

	var auths Collection
	err = r.lister.List(ctx, tenant, &auths, inOperatorConditions...)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching bundle_instance_auths")
	}

	return r.multipleFromEntities(auths)
}
