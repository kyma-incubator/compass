package labeldef

import (
	"context"
	"fmt"

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
	_, err = db.NamedExec(fmt.Sprintf("insert into %s (id, tenant_id, key, schema) values(:id, :tenantID, :key, :schemaJSON)", tableName), entity)
	if err != nil {
		return errors.Wrap(err, "while inserting Label Definition")
	}
	return nil
}
