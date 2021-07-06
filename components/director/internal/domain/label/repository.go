package label

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

const (
	tableName    string = "public.labels"
	tenantColumn string = "tenant_id"
)

var tableColumns = []string{"id", tenantColumn, "app_id", "runtime_id", "runtime_context_id", "key", "value"}

//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in model.Label) (Entity, error)
	FromEntity(in Entity) (model.Label, error)
}

type repository struct {
	upserter     repo.Upserter
	lister       repo.Lister
	deleter      repo.Deleter
	getter       repo.SingleGetter
	queryBuilder repo.QueryBuilder
	conv         Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		upserter:     repo.NewUpserter(resource.Label, tableName, tableColumns, []string{tenantColumn, "coalesce(app_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_context_id, '00000000-0000-0000-0000-000000000000')", "key"}, []string{"value"}),
		lister:       repo.NewLister(resource.Label, tableName, tenantColumn, tableColumns),
		deleter:      repo.NewDeleter(resource.Label, tableName, tenantColumn),
		getter:       repo.NewSingleGetter(resource.Label, tableName, tenantColumn, tableColumns),
		queryBuilder: repo.NewQueryBuilder(resource.Label, tableName, tenantColumn, []string{"runtime_id"}),
		conv:         conv,
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

func (r *repository) ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error) {
	var entities Collection
	if err := r.lister.List(ctx, tenant, &entities, repo.NewEqualCondition("key", key)); err != nil {
		return nil, err
	}

	var labels []*model.Label

	for _, entity := range entities {
		m, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labels = append(labels, &m)
	}

	return labels, nil
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
		repo.NewEqualCondition(tenantColumn, tenant),
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

func (r *repository) GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error) {
	subquery, args, err := r.queryBuilder.BuildQuery(tenantID, false,
		repo.NewEqualCondition("key",selectorKey),
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

func labelableObjectField(objectType model.LabelableObject) string {
	switch objectType {
	case model.ApplicationLabelableObject:
		return "app_id"
	case model.RuntimeLabelableObject:
		return "runtime_id"
	case model.RuntimeContextLabelableObject:
		return "runtime_context_id"
	}

	return ""
}
