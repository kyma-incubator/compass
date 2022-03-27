package systemauth

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.system_auths`

var (
	tableColumns = []string{"id", "tenant_id", "app_id", "runtime_id", "integration_system_id", "value"}
	tenantColumn = "tenant_id"
)

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in model.SystemAuth) (Entity, error)
	FromEntity(in Entity) (model.SystemAuth, error)
}

type repository struct {
	creator            repo.CreatorGlobal
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal
	updater            repo.UpdaterGlobal

	conv Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *repository {
	return &repository{
		creator:            repo.NewCreatorGlobal(resource.SystemAuth, tableName, tableColumns),
		singleGetter:       repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.SystemAuth, tableName, tableColumns),
		lister:             repo.NewListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		listerGlobal:       repo.NewListerGlobal(resource.SystemAuth, tableName, tableColumns),
		deleter:            repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		deleterGlobal:      repo.NewDeleterGlobal(resource.SystemAuth, tableName),
		updater:            repo.NewUpdaterWithEmbeddedTenant(resource.SystemAuth, tableName, []string{"value"}, tenantColumn, []string{"id"}),
		conv:               conv,
	}
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, item model.SystemAuth) error {
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	log.C(ctx).Debugf("Persisting SystemAuth entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

// GetByIDForObject missing godoc
func (r *repository) GetByIDForObject(ctx context.Context, tenant, id string, objType model.SystemAuthReferenceObjectType) (*model.SystemAuth, error) {
	column, err := referenceObjectField(objType)
	if err != nil {
		return nil, err
	}
	objTypeCond := repo.NewNotNullCondition(column)

	var entity Entity
	if err := r.singleGetter.Get(ctx, resource.SystemAuth, tenant, repo.Conditions{repo.NewEqualCondition("id", id), objTypeCond}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting SystemAuth entity to model")
	}

	return &itemModel, nil
}

// GetByIDGlobal missing godoc
func (r *repository) GetByIDGlobal(ctx context.Context, id string) (*model.SystemAuth, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting SystemAuth entity to model")
	}

	return &itemModel, nil
}

// GetByIDForObjectGlobal missing godoc
func (r *repository) GetByIDForObjectGlobal(ctx context.Context, id string, objType model.SystemAuthReferenceObjectType) (*model.SystemAuth, error) {
	column, err := referenceObjectField(objType)
	if err != nil {
		return nil, err
	}
	objTypeCond := repo.NewNotNullCondition(column)

	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id), objTypeCond}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting SystemAuth entity to model")
	}

	return &itemModel, nil
}

// GetByJSONValue missing godoc
func (r *repository) GetByJSONValue(ctx context.Context, value map[string]interface{}) (*model.SystemAuth, error) {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal")
	}
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewJSONCondition("value", string(valueBytes))}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting SystemAuth entity to model")
	}

	return &itemModel, nil
}

// ListForObject missing godoc
func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	objTypeFieldName, err := referenceObjectField(objectType)
	if err != nil {
		return nil, err
	}

	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition(objTypeFieldName, objectID),
	}

	err = r.lister.List(ctx, resource.SystemAuth, tenant, &entities, conditions...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListForObjectGlobal missing godoc
func (r *repository) ListForObjectGlobal(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	objTypeFieldName, err := referenceObjectField(objectType)
	if err != nil {
		return nil, err
	}

	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition(objTypeFieldName, objectID),
	}

	err = r.listerGlobal.ListGlobal(ctx, &entities, conditions...)
	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListGlobalWithConditions missing godoc
func (r *repository) ListGlobalWithConditions(ctx context.Context, conditions repo.Conditions) ([]model.SystemAuth, error) {
	var entities Collection

	if err := r.listerGlobal.ListGlobal(ctx, &entities, conditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

func (r *repository) multipleFromEntities(entities Collection) ([]model.SystemAuth, error) {
	items := make([]model.SystemAuth, 0, len(entities))

	for _, ent := range entities {
		m, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating system auth model from entity")
		}

		items = append(items, m)
	}

	return items, nil
}

// DeleteAllForObject missing godoc
func (r *repository) DeleteAllForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) error {
	objTypeFieldName, err := referenceObjectField(objectType)
	if err != nil {
		return err
	}
	if objectType == model.IntegrationSystemReference {
		return r.deleterGlobal.DeleteManyGlobal(ctx, repo.Conditions{repo.NewEqualCondition(objTypeFieldName, objectID)})
	}
	return r.deleter.DeleteMany(ctx, resource.SystemAuth, tenant, repo.Conditions{repo.NewEqualCondition(objTypeFieldName, objectID)})
}

// DeleteByIDForObject missing godoc
func (r *repository) DeleteByIDForObject(ctx context.Context, tenant, id string, objType model.SystemAuthReferenceObjectType) error {
	var objTypeCond repo.Condition

	column, err := referenceObjectField(objType)
	if err != nil {
		return err
	}
	objTypeCond = repo.NewNotNullCondition(column)

	return r.deleter.DeleteOne(ctx, resource.SystemAuth, tenant, repo.Conditions{repo.NewEqualCondition("id", id), objTypeCond})
}

// DeleteByIDForObjectGlobal missing godoc
func (r *repository) DeleteByIDForObjectGlobal(ctx context.Context, id string, objType model.SystemAuthReferenceObjectType) error {
	var objTypeCond repo.Condition

	column, err := referenceObjectField(objType)
	if err != nil {
		return err
	}
	objTypeCond = repo.NewNotNullCondition(column)

	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id), objTypeCond})
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, item *model.SystemAuth) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.updater.UpdateSingleGlobal(ctx, entity)
}

func referenceObjectField(objectType model.SystemAuthReferenceObjectType) (string, error) {
	switch objectType {
	case model.ApplicationReference:
		return "app_id", nil
	case model.RuntimeReference:
		return "runtime_id", nil
	case model.IntegrationSystemReference:
		return "integration_system_id", nil
	}

	return "", apperrors.NewInternalError("unsupported reference object type")
}
