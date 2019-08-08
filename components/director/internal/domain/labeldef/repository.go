package labeldef

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

const (
	tableName = "public.label_definitions"
)

type repo struct {
	conv Converter
}

func NewRepository(conv Converter) *repo {
	return &repo{conv: conv}
}

func (r *repo) Create(ctx context.Context, def model.LabelDefinition) error {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	entity, err := r.conv.ToEntity(def)
	if err != nil {
		return errors.Wrap(err, "while converting Label Definition to insert")
	}

	columns := []string{"id", "tenant_id", "key"}
	if entity.SchemaJSON.Valid {
		columns = append(columns, "schema")
	}
	values := r.prefixEveryWithColon(columns)

	_, err = db.NamedExec(fmt.Sprintf("insert into %s (%s) values(%s)", tableName, strings.Join(columns, ","), strings.Join(values, ",")), entity)
	if err != nil {
		return errors.Wrap(err, "while inserting Label Definition")
	}
	return nil
}

func (r *repo) prefixEveryWithColon(in []string) []string {
	out := make([]string, 0)
	for _, elem := range in {
		out = append(out, ":"+elem)
	}
	return out
}

func (r *repo) GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error) {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}
	dest := Entity{}

	q := fmt.Sprintf("select * from %s where tenant_id=$1 and key=$2 ", tableName)

	err = db.Get(&dest, q, tenant, key)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "while querying Label Definition")
	}

	ld, err := r.conv.FromEntity(dest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label Definition")
	}
	return &ld, nil
}

func (r *repo) Exists(ctx context.Context, tenant string, key string) (bool, error) {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	q := fmt.Sprintf("select 1 as exists from %s where tenant_id=$1 and key=$2 ", tableName)

	var count int
	err = db.Get(&count, q, tenant, key)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, errors.Wrap(err, "while querying Label Definition")
	}
	return true, nil
}

func (r *repo) List(ctx context.Context, tenant string) ([]model.LabelDefinition, error) {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var dest []Entity

	q := fmt.Sprintf("select * from %s where tenant_id=$1", tableName)

	if err = db.Select(&dest, q, tenant); err != nil {
		return nil, errors.Wrap(err, "while listing Label Definitions")
	}

	var out []model.LabelDefinition
	for _, entity := range dest {
		ld, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting Label Definition [key=%s]", entity.Key)
		}
		out = append(out, ld)
	}
	return out, nil
}

func (r *repo) Update(ctx context.Context, def model.LabelDefinition) error {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	entity, err := r.conv.ToEntity(def)
	if err != nil {
		return errors.Wrap(err, "while creating Label Definition entity from model")
	}

	stmt := fmt.Sprintf(`UPDATE %s SET "schema" = :schema WHERE "id" = :id`, tableName)
	result, err := db.NamedExec(stmt, entity)
	if err != nil {
		return errors.Wrap(err, "while updating Label Definition")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "while receiving affected rows in db")
	}

	if rowsAffected == 0 {
		return errors.New("no row was affected by query")
	}

	return nil
}

func (r *repo) DeleteByKey(ctx context.Context, tenant, key string) error {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE tenant_id=$1 AND key=$2`, tableName)
	result, err := db.Exec(stmt, tenant, key)
	if err != nil {
		return errors.Wrap(err, "while deleting the Label Definition entity from database")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "while receiving affected rows in db")
	}
	if rowsAffected < 1 {
		return errors.New("no rows were affected by query")
	}

	return nil
}
