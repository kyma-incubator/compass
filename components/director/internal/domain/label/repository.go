package label

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const tableName string = "public.labels"

var tableColumns = []string{"id", "tenant_id", "app_id", "runtime_id", "key", "value"}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.Label) (Entity, error)
	FromEntity(in Entity) (model.Label, error)
}

type repository struct {
	upserter repo.Upserter

	conv Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		upserter: repo.NewUpserter(tableName, tableColumns, []string{"tenant_id", "coalesce(app_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_id, '00000000-0000-0000-0000-000000000000')", "key"}, []string{"value"}),
		conv:     conv,
	}
}

func (r *repository) Upsert(ctx context.Context, label *model.Label) error {
	if label == nil {
		return errors.New("item can not be empty")
	}

	labelEntity, err := r.conv.ToEntity(*label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}

	return r.upserter.Upsert(ctx, labelEntity)
}

func (r *repository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE key = $1 AND %s = $2 AND tenant_id = $3`,
		strings.Join(tableColumns, ", "), tableName, labelableObjectField(objectType))

	var entity Entity
	err = persist.Get(&entity, stmt, key, objectID, tenant)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError(key)
		}
		return nil, errors.Wrap(err, "while getting Entity from DB")
	}

	labelModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label entity to model")
	}

	return &labelModel, nil
}

func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE  %s = $1 AND tenant_id = $2`,
		strings.Join(tableColumns, ", "), tableName, labelableObjectField(objectType))

	var entities []Entity
	err = persist.Select(&entities, stmt, objectID, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Labels from DB")
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
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE key = $1 AND tenant_id = $2`,
		strings.Join(tableColumns, ", "), tableName)

	var entities []Entity
	err = persist.Select(&entities, stmt, key, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Labels from DB")
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
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE key = $1 AND %s = $2 AND tenant_id = $3`, tableName, labelableObjectField(objectType))
	_, err = persist.Exec(stmt, key, objectID, tenant)

	return errors.Wrap(err, "while deleting the Label entity from database")
}

func (r *repository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1 AND tenant_id = $2`, tableName, labelableObjectField(objectType))
	_, err = persist.Exec(stmt, objectID, tenant)

	return errors.Wrapf(err, "while deleting all Label entities from database for %s %s", objectType, objectID)
}

func (r *repository) DeleteByKey(ctx context.Context, tenant string, key string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE key = $1 AND tenant_id = $2`, tableName)
	_, err = persist.Exec(stmt, key, tenant)
	if err != nil {
		return errors.Wrapf(err, `while deleting all Label entities from database with key "%s"`, key)
	}

	return nil
}

func labelableObjectField(objectType model.LabelableObject) string {
	switch objectType {
	case model.ApplicationLabelableObject:
		return "app_id"
	case model.RuntimeLabelableObject:
		return "runtime_id"
	}

	return ""
}
