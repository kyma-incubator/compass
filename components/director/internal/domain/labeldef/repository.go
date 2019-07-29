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
}

func NewRepository() *repo {
	return &repo{}
}

func (r *repo) Create(ctx context.Context, def model.LabelDefinition) error {
	db, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching transaction from context")
	}
	// _, err = txx.NamedExecContext(ctx, "insert into applications(id,tenant,name,description,labels) values (:id, :tenant, :name, :description, :labels)", appDTO)
	_, err = db.NamedExec(fmt.Sprintf("insert into %s (id, tenant_id, key, schema) values(:id, :tenant, :key, :schema)", tableName), def)
	if err != nil {
		return errors.Wrap(err, "while inserting Label Definition")
	}
	return nil
}
