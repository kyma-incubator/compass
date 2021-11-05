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

var (
	tableColumns       = []string{"id", tenantColumn, "app_id", "runtime_id", "runtime_context_id", "key", "value", "version"}
	updatableColumns   = []string{"value"}
	idColumns          = []string{"id"}
	versionedIDColumns = append(idColumns, "version")
)

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in model.Label) (*Entity, error)
	FromEntity(in Entity) (model.Label, error)
}

type repository struct {
	lister       repo.Lister
	listerGlobal repo.ListerGlobal
	deleter      repo.Deleter
	getter       repo.SingleGetter
	queryBuilder repo.QueryBuilder

	embeddedTenantLister       repo.Lister
	embeddedTenantDeleter      repo.Deleter
	embeddedTenantGetter       repo.SingleGetter
	embeddedTenantQueryBuilder repo.QueryBuilder

	creator                        repo.Creator
	globalCreator                  repo.CreatorGlobal
	updater                        repo.Updater
	versionedUpdater               repo.Updater
	embeddedTenantUpdater          repo.UpdaterGlobal
	versionedEmbeddedTenantUpdater repo.UpdaterGlobal
	conv                           Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		lister:       repo.NewLister(tableName, tableColumns),
		listerGlobal: repo.NewListerGlobal(resource.Label, tableName, tableColumns),
		deleter:      repo.NewDeleter(tableName),
		getter:       repo.NewSingleGetter(tableName, tableColumns),
		queryBuilder: repo.NewQueryBuilder(tableName, []string{"runtime_id"}),

		embeddedTenantLister:       repo.NewListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		embeddedTenantDeleter:      repo.NewDeleterWithEmbeddedTenant(tenantColumn, tableName),
		embeddedTenantGetter:       repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		embeddedTenantQueryBuilder: repo.NewQueryBuilderWithEmbeddedTenant(tableName, tenantColumn, []string{"runtime_id"}),

		creator:                        repo.NewCreator(tableName, tableColumns),
		globalCreator:                  repo.NewCreatorGlobal(resource.Label, tableName, tableColumns),
		updater:                        repo.NewUpdater(tableName, updatableColumns, idColumns),
		versionedUpdater:               repo.NewUpdater(tableName, updatableColumns, versionedIDColumns),
		embeddedTenantUpdater:          repo.NewUpdaterWithEmbeddedTenant(resource.Label, tableName, updatableColumns, tenantColumn, idColumns),
		versionedEmbeddedTenantUpdater: repo.NewUpdaterWithEmbeddedTenant(resource.Label, tableName, updatableColumns, tenantColumn, versionedIDColumns),
		conv:                           conv,
	}
}

// Upsert missing godoc
func (r *repository) Upsert(ctx context.Context, tenant string, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	l, err := r.GetByKey(ctx, tenant, label.ObjectType, label.ObjectID, label.Key)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return r.Create(ctx, tenant, label)
		}
		return err
	}

	l.Value = label.Value
	labelEntity, err := r.conv.ToEntity(*l)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}
	if label.ObjectType == model.TenantLabelableObject {
		return r.embeddedTenantUpdater.UpdateSingleWithVersionGlobal(ctx, labelEntity)
	}
	return r.updater.UpdateSingleWithVersion(ctx, label.ObjectType.GetResourceType(), tenant, labelEntity)
}

// UpdateWithVersion missing godoc
func (r *repository) UpdateWithVersion(ctx context.Context, tenant string, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}
	labelEntity, err := r.conv.ToEntity(*label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}
	if label.ObjectType == model.TenantLabelableObject {
		return r.embeddedTenantUpdater.UpdateSingleWithVersionGlobal(ctx, labelEntity)
	}
	return r.versionedUpdater.UpdateSingleWithVersion(ctx, label.ObjectType.GetResourceType(), tenant, labelEntity)
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, tenant string, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	labelEntity, err := r.conv.ToEntity(*label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}

	if label.ObjectType == model.TenantLabelableObject {
		return r.globalCreator.Create(ctx, labelEntity)
	}

	return r.creator.Create(ctx, label.ObjectType.GetResourceType(), tenant, labelEntity)
}

// GetByKey missing godoc
func (r *repository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error) {
	getter := r.getter
	if objectType == model.TenantLabelableObject {
		getter = r.embeddedTenantGetter
	}

	var entity Entity
	if err := getter.Get(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition("key", key), repo.NewEqualCondition(labelableObjectField(objectType), objectID)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	labelModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label entity to model")
	}

	return &labelModel, nil
}

// ListForObject missing godoc
func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	var entities Collection

	typeCondition := repo.NewEqualCondition(labelableObjectField(objectType), objectID)
	conditions := []repo.Condition{typeCondition}

	lister := r.lister
	if objectType == model.TenantLabelableObject {
		lister = r.embeddedTenantLister
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.ApplicationLabelableObject)))
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.RuntimeContextLabelableObject)))
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.RuntimeLabelableObject)))
	}

	if err := lister.List(ctx, objectType.GetResourceType(), tenant, &entities, conditions...); err != nil {
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

// ListByKey missing godoc
func (r *repository) ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error) {
	var entities Collection
	if err := r.lister.List(ctx, resource.Label, tenant, &entities, repo.NewEqualCondition("key", key)); err != nil {
		return nil, err
	}

	labels := make([]*model.Label, 0, len(entities))

	for _, entity := range entities {
		m, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labels = append(labels, &m)
	}

	return labels, nil
}

// ListGlobalByKeyAndObjects lists resources of objectType across tenants (global) which match the given objectIDs and labeled with the provided key
func (r *repository) ListGlobalByKeyAndObjects(ctx context.Context, objectType model.LabelableObject, objectIDs []string, key string) ([]*model.Label, error) {
	var entities Collection
	if err := r.listerGlobal.ListGlobal(ctx, &entities, repo.NewEqualCondition("key", key), repo.NewInConditionForStringValues(labelableObjectField(objectType), objectIDs)); err != nil {
		return nil, err
	}

	labels := make([]*model.Label, 0, len(entities))

	for _, entity := range entities {
		m, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labels = append(labels, &m)
	}

	return labels, nil
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error {
	deleter := r.deleter
	if objectType == model.TenantLabelableObject {
		deleter = r.embeddedTenantDeleter
	}
	return deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition("key", key), repo.NewEqualCondition(labelableObjectField(objectType), objectID)})
}

// DeleteAll missing godoc
func (r *repository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	deleter := r.deleter
	if objectType == model.TenantLabelableObject {
		deleter = r.embeddedTenantDeleter
	}
	return deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition(labelableObjectField(objectType), objectID)})
}

// DeleteByKeyNegationPattern missing godoc
func (r *repository) DeleteByKeyNegationPattern(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labelKeyPattern string) error {
	deleter := r.deleter
	if objectType == model.TenantLabelableObject {
		deleter = r.embeddedTenantDeleter
	}

	return deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, repo.Conditions{
		repo.NewEqualCondition(labelableObjectField(objectType), objectID),
		repo.NewEqualCondition(tenantColumn, tenant),
		repo.NewNotRegexConditionString("key", labelKeyPattern),
	})
}

// DeleteByKey missing godoc
func (r *repository) DeleteByKey(ctx context.Context, tenant string, key string) error {
	return r.deleter.DeleteMany(ctx, resource.Label, tenant, repo.Conditions{repo.NewEqualCondition("key", key)})
}

// GetRuntimesIDsByStringLabel missing godoc
func (r *repository) GetRuntimesIDsByStringLabel(ctx context.Context, tenantID, key, value string) ([]string, error) {
	var entities Collection
	if err := r.lister.List(ctx, resource.RuntimeLabel, tenantID, &entities,
		repo.NewEqualCondition("key", key),
		repo.NewJSONArrMatchAnyStringCondition("value", value),
		repo.NewNotNullCondition("runtime_id")); err != nil {
		return nil, err
	}

	matchedRtmsIDs := make([]string, 0, len(entities))
	for _, entity := range entities {
		matchedRtmsIDs = append(matchedRtmsIDs, entity.RuntimeID.String)
	}
	return matchedRtmsIDs, nil
}

// GetScenarioLabelsForRuntimes missing godoc
func (r *repository) GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error) {
	if len(runtimesIDs) == 0 {
		return nil, apperrors.NewInvalidDataError("cannot execute query without runtimeIDs")
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition("key", model.ScenariosKey),
		repo.NewInConditionForStringValues("runtime_id", runtimesIDs),
	}

	var labels Collection
	err := r.lister.List(ctx, resource.RuntimeLabel, tenantID, &labels, conditions...)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes scenarios")
	}

	labelModels := make([]model.Label, 0, len(labels))
	for _, label := range labels {
		labelModel, err := r.conv.FromEntity(label)
		if err != nil {
			return nil, errors.Wrap(err, "while converting label entity to model")
		}
		labelModels = append(labelModels, labelModel)
	}
	return labelModels, nil
}

// GetRuntimeScenariosWhereLabelsMatchSelector missing godoc
func (r *repository) GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error) {
	subquery, args, err := r.queryBuilder.BuildQuery(resource.RuntimeLabel, tenantID, false,
		repo.NewEqualCondition("key", selectorKey),
		repo.NewJSONArrMatchAnyStringCondition("value", selectorValue),
		repo.NewNotNullCondition("runtime_id"))

	if err != nil {
		return nil, errors.Wrap(err, "while building subquery")
	}

	var labels Collection
	if err := r.lister.List(ctx, resource.RuntimeLabel, tenantID, &labels, repo.NewEqualCondition("key", "scenarios"), repo.NewInConditionForSubQuery("runtime_id", subquery, args)); err != nil {
		return nil, err
	}

	labelModels := make([]model.Label, 0, len(labels))
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
	case model.TenantLabelableObject:
		return "tenant_id"
	}

	return ""
}
