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
	tableColumns       = []string{"id", tenantColumn, "app_id", "runtime_id", "runtime_context_id", "app_template_id", "key", "value", "version"}
	updatableColumns   = []string{"value"}
	idColumns          = []string{"id"}
	versionedIDColumns = append(idColumns, "version")
)

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in *model.Label) (*Entity, error)
	FromEntity(in *Entity) (*model.Label, error)
}

type repository struct {
	lister        repo.Lister
	listerGlobal  repo.ListerGlobal
	deleter       repo.Deleter
	deleterGlobal repo.DeleterGlobal
	getter        repo.SingleGetter
	getterGlobal  repo.SingleGetterGlobal

	embeddedTenantLister  repo.Lister
	embeddedTenantDeleter repo.Deleter
	embeddedTenantGetter  repo.SingleGetter

	creator                        repo.Creator
	globalCreator                  repo.CreatorGlobal
	updater                        repo.Updater
	versionedUpdater               repo.Updater
	updaterGlobal                  repo.UpdaterGlobal
	embeddedTenantUpdater          repo.UpdaterGlobal
	versionedEmbeddedTenantUpdater repo.UpdaterGlobal
	conv                           Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		lister:        repo.NewLister(tableName, tableColumns),
		listerGlobal:  repo.NewListerGlobal(resource.Label, tableName, tableColumns),
		deleter:       repo.NewDeleter(tableName),
		deleterGlobal: repo.NewDeleterGlobal(resource.Label, tableName),
		getter:        repo.NewSingleGetter(tableName, tableColumns),
		getterGlobal:  repo.NewSingleGetterGlobal(resource.Label, tableName, tableColumns),

		embeddedTenantLister:  repo.NewListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		embeddedTenantDeleter: repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		embeddedTenantGetter:  repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),

		creator:                        repo.NewCreator(tableName, tableColumns),
		globalCreator:                  repo.NewCreatorGlobal(resource.Label, tableName, tableColumns),
		updater:                        repo.NewUpdater(tableName, updatableColumns, idColumns),
		versionedUpdater:               repo.NewUpdater(tableName, updatableColumns, versionedIDColumns),
		updaterGlobal:                  repo.NewUpdaterGlobal(resource.Label, tableName, updatableColumns, idColumns),
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
	labelEntity, err := r.conv.ToEntity(l)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}
	if label.ObjectType == model.TenantLabelableObject {
		return r.embeddedTenantUpdater.UpdateSingleWithVersionGlobal(ctx, labelEntity)
	}
	if label.ObjectType == model.AppTemplateLabelableObject {
		return r.updaterGlobal.UpdateSingleWithVersionGlobal(ctx, labelEntity)
	}

	return r.updater.UpdateSingleWithVersion(ctx, label.ObjectType.GetResourceType(), tenant, labelEntity)
}

// UpdateWithVersion missing godoc
func (r *repository) UpdateWithVersion(ctx context.Context, tenant string, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}
	labelEntity, err := r.conv.ToEntity(label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}
	if label.ObjectType == model.TenantLabelableObject {
		return r.versionedEmbeddedTenantUpdater.UpdateSingleWithVersionGlobal(ctx, labelEntity)
	}
	return r.versionedUpdater.UpdateSingleWithVersion(ctx, label.ObjectType.GetResourceType(), tenant, labelEntity)
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, tenant string, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	labelEntity, err := r.conv.ToEntity(label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}

	if label.ObjectType == model.TenantLabelableObject || label.ObjectType == model.AppTemplateLabelableObject { // ????????
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

	conds := repo.Conditions{repo.NewEqualCondition("key", key)}
	if objectType != model.TenantLabelableObject {
		conds = append(conds, repo.NewEqualCondition(labelableObjectField(objectType), objectID))
	}

	var entity Entity

	if objectType == model.AppTemplateLabelableObject {
		if err := r.getterGlobal.GetGlobal(ctx, conds, repo.NoOrderBy, &entity); err != nil {
			return nil, err
		}
	} else {
		if err := getter.Get(ctx, objectType.GetResourceType(), tenant, conds, repo.NoOrderBy, &entity); err != nil {
			return nil, err
		}
	}

	labelModel, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label entity to model")
	}

	return labelModel, nil
}

// ListForObject missing godoc
func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	var entities Collection

	var conditions []repo.Condition

	lister := r.lister
	if objectType == model.TenantLabelableObject {
		lister = r.embeddedTenantLister
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.ApplicationLabelableObject)))
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.RuntimeContextLabelableObject)))
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.RuntimeLabelableObject)))
		conditions = append(conditions, repo.NewNullCondition(labelableObjectField(model.AppTemplateLabelableObject)))
	} else {
		conditions = append(conditions, repo.NewEqualCondition(labelableObjectField(objectType), objectID))
	}

	if objectType == model.AppTemplateLabelableObject{
		if err := r.listerGlobal.ListGlobal(ctx, &entities, conditions...); err != nil {
			return nil, err
		}
	} else {
		if err := lister.List(ctx, objectType.GetResourceType(), tenant, &entities, conditions...); err != nil {
			return nil, err
		}
	}

	labelsMap := make(map[string]*model.Label)

	for _, entity := range entities {
		m, err := r.conv.FromEntity(&entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labelsMap[m.Key] = m
	}

	return labelsMap, nil
}

// ListByKey missing godoc
func (r *repository) ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error) {
	var entities Collection
	if err := r.lister.List(ctx, resource.Label, tenant, &entities, repo.NewEqualCondition("key", key)); err != nil {
		return nil, err
	}
	return r.multipleFromEntity(entities)
}

// ListGlobalByKey lists all labels which are labeled with the provided key across tenants (global)
func (r *repository) ListGlobalByKey(ctx context.Context, key string) ([]*model.Label, error) {
	var entities Collection

	if err := r.listerGlobal.ListGlobal(ctx, &entities, repo.NewEqualCondition("key", key)); err != nil {
		return nil, err
	}

	return r.multipleFromEntity(entities)
}

// ListGlobalByKeyAndObjects lists resources of objectType across tenants (global) which match the given objectIDs and labeled with the provided key
func (r *repository) ListGlobalByKeyAndObjects(ctx context.Context, objectType model.LabelableObject, objectIDs []string, key string) ([]*model.Label, error) {
	var entities Collection
	if err := r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &entities, repo.NewEqualCondition("key", key), repo.NewInConditionForStringValues(labelableObjectField(objectType), objectIDs)); err != nil {
		return nil, err
	}
	return r.multipleFromEntity(entities)
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error {
	deleter := r.deleter
	conds := repo.Conditions{repo.NewEqualCondition("key", key)}
	if objectType == model.TenantLabelableObject {
		deleter = r.embeddedTenantDeleter
	} else {
		conds = append(conds, repo.NewEqualCondition(labelableObjectField(objectType), objectID))
	}
	if objectType == model.AppTemplateLabelableObject {
		return r.deleterGlobal.DeleteManyGlobal(ctx, conds)
	}
	return deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, conds)
}

// DeleteAll missing godoc
func (r *repository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	deleter := r.deleter
	conds := repo.Conditions{}
	if objectType == model.TenantLabelableObject {
		deleter = r.embeddedTenantDeleter
	} else {
		conds = append(conds, repo.NewEqualCondition(labelableObjectField(objectType), objectID))
	}
	if objectType == model.AppTemplateLabelableObject {
		return r.deleterGlobal.DeleteManyGlobal(ctx, conds)
	}
	return deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, conds)
}

// DeleteByKeyNegationPattern missing godoc
func (r *repository) DeleteByKeyNegationPattern(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labelKeyPattern string) error {
	deleter := r.deleter
	conds := repo.Conditions{repo.NewNotRegexConditionString("key", labelKeyPattern)}
	if objectType == model.TenantLabelableObject {
		deleter = r.embeddedTenantDeleter
	} else {
		conds = append(conds, repo.NewEqualCondition(labelableObjectField(objectType), objectID))
	}
	if objectType == model.AppTemplateLabelableObject {
		return r.deleterGlobal.DeleteManyGlobal(ctx, conds)
	}
	return deleter.DeleteMany(ctx, objectType.GetResourceType(), tenant, conds)
}

// DeleteByKey missing godoc
func (r *repository) DeleteByKey(ctx context.Context, tenant string, key string) error {
	return r.deleter.DeleteMany(ctx, resource.Label, tenant, repo.Conditions{repo.NewEqualCondition("key", key)})
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
		labelModel, err := r.conv.FromEntity(&label)
		if err != nil {
			return nil, errors.Wrap(err, "while converting label entity to model")
		}
		labelModels = append(labelModels, *labelModel)
	}
	return labelModels, nil
}

func (r *repository) multipleFromEntity(entities []Entity) ([]*model.Label, error) {
	labels := make([]*model.Label, 0, len(entities))

	for _, entity := range entities {
		m, err := r.conv.FromEntity(&entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labels = append(labels, m)
	}

	return labels, nil
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
	case model.AppTemplateLabelableObject:
		return "app_template_id"
	}

	return ""
}
