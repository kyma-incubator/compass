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

var tableColumns = []string{"id", "tenant_id", "app_id", "runtime_id", "integration_system_id", "value"}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.SystemAuth) (Entity, error)
	FromEntity(in Entity) (model.SystemAuth, error)
}

type pgRepository struct {
	*repo.Creator
	*repo.Lister
	*repo.Deleter

	conv Converter
}

func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		Creator: repo.NewCreator(tableName, tableColumns),
		Lister:  repo.NewLister(tableName, "tenant_id", tableColumns),
		Deleter: repo.NewDeleter(tableName, "tenant_id"),
		conv:    conv,
	}
}

func (r *pgRepository) Create(ctx context.Context, item model.SystemAuth) error {
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.Creator.Create(ctx, entity)
}

func (r *pgRepository) ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	objType, err := referenceObjectField(objectType)
	if err != nil {
		return nil, err
	}

	var entities Collection
	if err := r.Lister.List(ctx, tenant, &entities, fmt.Sprintf("%s = %s", objType, pq.QuoteLiteral(objectID))); err != nil {
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

func (r *pgRepository) Delete(ctx context.Context, tenant string, id string, objectType model.SystemAuthReferenceObjectType) error {
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
