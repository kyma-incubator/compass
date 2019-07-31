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
	def, err := r.GetByKey(ctx, tenant, key)
	if err != nil {
		return false, err
	}
	if def != nil {
		return true, nil
	}
	return false, nil
}
