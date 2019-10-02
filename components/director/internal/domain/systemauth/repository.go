package systemauth

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/lib/pq"
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
	repo.Creator
	repo.SingleGetter
	repo.Lister
	repo.Deleter

	conv Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		Creator:      repo.NewCreator(tableName, tableColumns),
		SingleGetter: repo.NewSingleGetter(tableName, tenantColumn, tableColumns),
		Lister:       repo.NewLister(tableName, tenantColumn, tableColumns),
		Deleter:      repo.NewDeleter(tableName, tenantColumn),
		conv:         conv,
	}
}

func (r *repository) Create(ctx context.Context, item model.SystemAuth) error {
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.Creator.Create(ctx, entity)
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.SystemAuth, error) {
	var entity Entity
	if err := r.SingleGetter.Get(ctx, tenant, repo.Conditions{{Field: "id", Val: id}}, &entity); err != nil {
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
	if err := r.Lister.List(ctx, tenant, &entities, fmt.Sprintf("%s = %s", objTypeFieldName, pq.QuoteLiteral(objectID))); err != nil {
		return nil, err
	}

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

	return r.Deleter.DeleteMany(ctx, tenant, repo.Conditions{{Field: objTypeFieldName, Val: objectID}})
}

func (r *repository) Delete(ctx context.Context, tenant string, id string) error {
	return r.Deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
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

	return "", errors.New("unsupported reference object type")
}
