package label

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const tableName string = `"public"."labels"`
const fields string = `"id", "tenant_id", "key", "value", "app_id", "runtime_id"`

type dbRepository struct {
}

func NewRepository() *dbRepository {
	return &dbRepository{}
}

func (r *dbRepository) Upsert(ctx context.Context, label *model.Label) error {
	if label == nil {
		return errors.New("Item cannot be empty")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	entity, err := EntityFromModel(label)
	if err != nil {
		return errors.Wrap(err, "while creating Label entity from model")
	}

	stmt := fmt.Sprintf(`INSERT INTO %s (id, tenant_id, key, value, app_id, runtime_id) VALUES (:id, :tenant_id, :key, :value, :app_id, :runtime_id)
		ON CONFLICT (id) DO UPDATE SET
    		key = EXCLUDED.key,
    		value = EXCLUDED.value,
    		app_id = EXCLUDED.app_id,
    		runtime_id = EXCLUDED.runtime_id
		`, tableName)

	_, err = persist.NamedExec(stmt, entity)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == persistence.UniqueViolation {
			return errors.Wrap(pqErr, "unique Violation error:")
		}

		return errors.Wrap(err, "while upserting the Label entity to database")
	}

	return nil
}

func (r *dbRepository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE "key" = $1 AND "%s" = $2 AND "tenant_id" = $3`,
		fields, tableName, r.objectField(objectType))

	var entity Entity
	err = persist.Get(&entity, stmt, key, objectID, tenant)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "while getting Entity from DB")
		}

		return nil, fmt.Errorf("label '%s' not found", key) //TODO: Return own type for Not found error
	}

	labelModel, err := entity.ToModel()
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label entity to model")
	}

	return &labelModel, nil
}

func (r *dbRepository) List(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE  "%s" = $1 AND "tenant_id" = $2`,
		fields, tableName, r.objectField(objectType))

	var entities []Entity
	err = persist.Select(&entities, stmt, objectID, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Labels from DB")
	}

	labelsMap := make(map[string]*model.Label)

	for _, entity := range entities {
		m, err := entity.ToModel()
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labelsMap[m.Key] = &m
	}

	return labelsMap, nil
}

func (r *dbRepository) Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE "key" = $1 AND "%s" = $2 AND "tenant_id" = $3`, tableName, r.objectField(objectType))
	_, err = persist.Exec(stmt, key, objectID, tenant)

	return errors.Wrap(err, "while deleting the Label entity from database")
}

func (r *dbRepository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE "%s" = $1 AND "tenant_id" = $2`, tableName, r.objectField(objectType))
	_, err = persist.Exec(stmt, objectID, tenant)

	return errors.Wrapf(err, "while deleting all Label entities from database for %s %s", objectType, objectID)
}

func (r *dbRepository) objectField(objectType model.LabelableObject) string {
	switch objectType {
	case model.ApplicationLabelableObject:
		return "app_id"
	case model.RuntimeLabelableObject:
		return "runtime_id"
	}

	return ""
}
