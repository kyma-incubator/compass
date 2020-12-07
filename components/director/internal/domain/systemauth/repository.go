package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const tableName string = `public.system_auths`

var (
	tableColumns = []string{"id", "tenant_id", "app_id", "runtime_id", "integration_system_id", "value"}
	tenantColumn = "tenant_id"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.SystemAuth) (Entity, error)
	FromEntity(in Entity) (model.SystemAuth, error)
}

type repository struct {
	creator            repo.Creator
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal

	conv Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		creator:            repo.NewCreator(resource.SystemAuth, tableName, tableColumns),
		singleGetter:       repo.NewSingleGetter(resource.SystemAuth, tableName, tenantColumn, tableColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.SystemAuth, tableName, tableColumns),
		lister:             repo.NewLister(resource.SystemAuth, tableName, tenantColumn, tableColumns),
		listerGlobal:       repo.NewListerGlobal(resource.SystemAuth, tableName, tableColumns),
		deleter:            repo.NewDeleter(resource.SystemAuth, tableName, tenantColumn),
		deleterGlobal:      repo.NewDeleterGlobal(resource.SystemAuth, tableName),
		conv:               conv,
	}
}

func (r *repository) Create(ctx context.Context, item model.SystemAuth) error {
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	log.C(ctx).Debugf("Persisting SystemAuth entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.SystemAuth, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting SystemAuth entity to model")
	}

	return &itemModel, nil
}

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

func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	objTypeFieldName, err := referenceObjectField(objectType)
	if err != nil {
		return nil, err
	}

	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition(objTypeFieldName, objectID),
	}

	err = r.lister.List(ctx, tenant, &entities, conditions...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

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

func (r *repository) multipleFromEntities(entities Collection) ([]model.SystemAuth, error) {

	var items []model.SystemAuth

	for _, ent := range entities {
		m, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating system auth model from entity")
		}

		items = append(items, m)
	}

	return items, nil
}

func (r *repository) DeleteAllForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) error {
	objTypeFieldName, err := referenceObjectField(objectType)
	if err != nil {
		return err
	}
	if objectType == model.IntegrationSystemReference {
		return r.deleterGlobal.DeleteManyGlobal(ctx, repo.Conditions{repo.NewEqualCondition(objTypeFieldName, objectID)})
	}
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{repo.NewEqualCondition(objTypeFieldName, objectID)})
}

func (r *repository) DeleteByIDForObject(ctx context.Context, tenant, id string, objType model.SystemAuthReferenceObjectType) error {
	var objTypeCond repo.Condition

	column, err := referenceObjectField(objType)
	if err != nil {
		return err
	}
	objTypeCond = repo.NewNotNullCondition(column)

	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id), objTypeCond})
}

func (r *repository) DeleteByIDForObjectGlobal(ctx context.Context, id string, objType model.SystemAuthReferenceObjectType) error {
	var objTypeCond repo.Condition

	column, err := referenceObjectField(objType)
	if err != nil {
		return err
	}
	objTypeCond = repo.NewNotNullCondition(column)

	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id), objTypeCond})
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
