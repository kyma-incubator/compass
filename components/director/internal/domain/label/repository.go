package label

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	BundleInstanceAuthScenariosViewName        = `public.bundle_instance_auths_scenarios_labels`
	tableName                           string = "public.labels"
	tenantColumn                        string = "tenant_id"
	idColumn                            string = "id"
)

var tableColumns = []string{idColumn, tenantColumn, "app_id", "runtime_id", "bundle_instance_auth_id", "runtime_context_id", "key", "value"}

//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in model.Label) (Entity, error)
	FromEntity(in Entity) (model.Label, error)
	MultipleFromEntities(entities Collection) ([]model.Label, error)
}

type repository struct {
	upserter                               repo.Upserter
	lister                                 repo.Lister
	deleter                                repo.Deleter
	conv                                   Converter
	bundleInstanceAuthScenarioQueryBuilder repo.QueryBuilder
	getter                                 repo.SingleGetter
	queryBuilder                           repo.QueryBuilder
}

func NewRepository(conv Converter) *repository {
	return &repository{
		upserter:                               repo.NewUpserter(resource.Label, tableName, tableColumns, []string{tenantColumn, "coalesce(app_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_id, '00000000-0000-0000-0000-000000000000')", "coalesce(bundle_instance_auth_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_context_id, '00000000-0000-0000-0000-000000000000')", "key"}, []string{"value"}),
		lister:                                 repo.NewLister(resource.Label, tableName, tenantColumn, tableColumns),
		deleter:                                repo.NewDeleter(resource.Label, tableName, tenantColumn),
		conv:                                   conv,
		bundleInstanceAuthScenarioQueryBuilder: repo.NewQueryBuilder(resource.Label, BundleInstanceAuthScenariosViewName, tenantColumn, []string{"label_id"}),
		getter:                                 repo.NewSingleGetter(resource.Label, tableName, tenantColumn, tableColumns),
		queryBuilder:                           repo.NewQueryBuilder(resource.Label, tableName, tenantColumn, []string{"runtime_id"}),
	}
}

func (r *repository) Upsert(ctx context.Context, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	labelEntity, err := r.conv.ToEntity(*label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}

	return r.upserter.Upsert(ctx, labelEntity)
}

func (r *repository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error) {
	var entity Entity
	if err := r.getter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("key", key), repo.NewEqualCondition(labelableObjectField(objectType), objectID)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	labelModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label entity to model")
	}

	return &labelModel, nil
}

func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	var entities Collection
	if err := r.lister.List(ctx, tenant, &entities, repo.NewEqualCondition(labelableObjectField(objectType), objectID)); err != nil {
		return nil, err
	}

	labelsMap := make(map[string]*model.Label)

	for _, entity := range entities {
		m, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labelsMap[m.Key] = &m
	}

	return labelsMap, nil
}

func (r *repository) ListByKey(ctx context.Context, tenant, key string) ([]model.Label, error) {
	var entities Collection
	if err := r.lister.List(ctx, tenant, &entities, repo.NewEqualCondition("key", key)); err != nil {
		return nil, err
	}

	return r.conv.MultipleFromEntities(entities)
}

func (r *repository) Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("key", key), repo.NewEqualCondition(labelableObjectField(objectType), objectID)})
}

func (r *repository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{repo.NewEqualCondition(labelableObjectField(objectType), objectID)})
}

func (r *repository) DeleteByKeyNegationPattern(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labelKeyPattern string) error {
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{
		repo.NewEqualCondition(labelableObjectField(objectType), objectID),
		repo.NewNotRegexConditionString("key", labelKeyPattern),
	})
}

func (r *repository) DeleteByKey(ctx context.Context, tenant string, key string) error {
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{repo.NewEqualCondition("key", key)})
}

func (r *repository) GetRuntimesIDsByStringLabel(ctx context.Context, tenantID, key, value string) ([]string, error) {
	var entities Collection
	if err := r.lister.List(ctx, tenantID, &entities,
		repo.NewEqualCondition("key", key),
		repo.NewJSONArrMatchAnyStringCondition("value", value),
		repo.NewNotNullCondition("runtime_id")); err != nil {
		return nil, err
	}

	var matchedRtmsIDs []string
	for _, entity := range entities {
		matchedRtmsIDs = append(matchedRtmsIDs, entity.RuntimeID.String)
	}
	return matchedRtmsIDs, nil
}

func (r *repository) GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error) {
	if len(runtimesIDs) == 0 {
		return nil, apperrors.NewInvalidDataError("cannot execute query without runtimeIDs")
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition("key", model.ScenariosKey),
		repo.NewInConditionForStringValues("runtime_id", runtimesIDs),
	}

	var labels Collection
	err := r.lister.List(ctx, tenantID, &labels, conditions...)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes scenarios")
	}

	return r.conv.MultipleFromEntities(labels)
}

func (r *repository) GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error) {
	subquery, args, err := r.queryBuilder.BuildQuery(tenantID, false,
		repo.NewEqualCondition("key", selectorKey),
		repo.NewJSONArrMatchAnyStringCondition("value", selectorValue),
		repo.NewNotNullCondition("runtime_id"))

	if err != nil {
		return nil, errors.Wrap(err, "while building subquery")
	}

	var labels Collection
	if err := r.lister.List(ctx, tenantID, &labels, repo.NewEqualCondition("key", "scenarios"), repo.NewInConditionForSubQuery("runtime_id", subquery, args)); err != nil {
		return nil, err
	}

	var labelModels []model.Label
	for _, label := range labels {

		labelModel, err := r.conv.FromEntity(label)
		if err != nil {
			return nil, errors.Wrap(err, "while converting label entity to model")
		}
		labelModels = append(labelModels, labelModel)
	}
	return labelModels, nil
}

func (r *repository) GetBundleInstanceAuthsScenarioLabels(ctx context.Context, tenant, appId, runtimeId string) ([]model.Label, error) {
	subqueryConditions := repo.Conditions{
		repo.NewEqualCondition("app_id", appId),
		repo.NewEqualCondition("runtime_id", runtimeId),
	}

	subquery, args, err := r.bundleInstanceAuthScenarioQueryBuilder.BuildQuery(tenant, false, subqueryConditions...)
	if err != nil {
		return nil, err
	}

	inOperatorConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, subquery, args),
	}

	var labels Collection
	err = r.lister.List(ctx, tenant, &labels, inOperatorConditions...)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching bundle_instance_auth scenario labels")
	}

	return r.conv.MultipleFromEntities(labels)
}

func (r *repository) ListByObjectTypeAndMatchAnyScenario(ctx context.Context, tenantId string, objectType model.LabelableObject, scenarios []string) ([]model.Label, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("key", model.ScenariosKey),
		repo.NewNotNullCondition(labelableObjectField(objectType)),
		repo.NewJSONArrMatchAnyStringCondition("value", scenarios...),
	}

	var labels Collection
	err := r.lister.List(ctx, tenantId, &labels, conditions...)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes scenarios")
	}

	return r.conv.MultipleFromEntities(labels)
}

func labelableObjectField(objectType model.LabelableObject) string {
	switch objectType {
	case model.ApplicationLabelableObject:
		return "app_id"
	case model.RuntimeLabelableObject:
		return "runtime_id"
	case model.RuntimeContextLabelableObject:
		return "runtime_context_id"
	case model.BundleInstanceAuthLabelableObject:
		return "bundle_instance_auth_id"
	}

	return ""
}
