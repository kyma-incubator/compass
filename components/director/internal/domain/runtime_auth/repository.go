package runtime_auth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"
)

const tableName string = `public.runtime_auths`

var tableColumns = []string{"id", "tenant_id", "runtime_id", "api_def_id", "value"}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.RuntimeAuth) (Entity, error)
	FromEntity(in Entity) (model.RuntimeAuth, error)
}

type pgRepository struct {
	*repo.SingleGetter
	*repo.Lister
	*repo.Upserter
	*repo.Deleter

	conv Converter
}

func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		SingleGetter: repo.NewSingleGetter(tableName, "tenant_id", tableColumns),
		Lister:       repo.NewLister(tableName, "tenant_id", tableColumns),
		Upserter:     repo.NewUpserter(tableName, tableColumns, []string{"tenant_id", "runtime_id", "api_def_id"}, []string{"value"}),
		Deleter:      repo.NewDeleter(tableName, "tenant_id"),
		conv:         conv,
	}
}

func (r *pgRepository) Get(ctx context.Context, tenant string, apiID string, runtimeID string) (*model.RuntimeAuth, error) {
	var ent Entity
	if err := r.SingleGetter.Get(ctx, tenant, repo.Conditions{{Field: "runtime_id", Val: runtimeID}, {Field: "api_def_id", Val: apiID}}, &ent); err != nil {
		return nil, err
	}

	runtimeAuthModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating runtime auth model from entity")
	}

	return &runtimeAuthModel, nil
}

func (r *pgRepository) GetOrDefault(ctx context.Context, tenant string, apiID string, runtimeID string) (*model.RuntimeAuth, error) {
	var ent Entity

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ra.id, $2 AS api_def_id,
	COALESCE(ra.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value
	FROM (SELECT * FROM runtimes WHERE id = $3) AS r
	LEFT OUTER JOIN (SELECT * FROM runtime_auths
	WHERE api_def_id = $2 AND runtime_id = $3 AND tenant_id = $1) AS ra ON ra.runtime_id = r.id`

	err = persist.Get(&ent, stmt, tenant, apiID, runtimeID)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "while getting runtime auth or default runtime auth from DB")
	}

	runtimeAuthModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating runtime auth model from entity")
	}

	return &runtimeAuthModel, nil
}

type RuntimeAuthCollection []Entity

func (r RuntimeAuthCollection) Len() int {
	return len(r)
}

func (r *pgRepository) ListForAllRuntimes(ctx context.Context, tenant string, apiID string) ([]model.RuntimeAuth, error) {
	var runtimeAuthsCollection RuntimeAuthCollection

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ra.id, $2 AS api_def_id,
			coalesce(ra.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value
    		FROM (SELECT * FROM runtime_auths WHERE api_def_id = $2 AND tenant_id = $1) AS ra
    		RIGHT OUTER JOIN runtimes AS r ON ra.runtime_id = r.id WHERE r.tenant_id = $1`

	err = persist.Select(&runtimeAuthsCollection, stmt, tenant, apiID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching list of objects from DB")
	}

	var items []model.RuntimeAuth

	for _, ent := range runtimeAuthsCollection {
		m, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating runtime auth model from entity")
		}

		items = append(items, m)
	}

	return items, nil
}

func (r *pgRepository) Upsert(ctx context.Context, item model.RuntimeAuth) error {
	runtimeEnt, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime auth entity from model")
	}

	return r.Upserter.Upsert(ctx, runtimeEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant string, apiID string, runtimeID string) error {
	return r.Deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "api_def_id", Val: apiID}, {Field: "runtime_id", Val: runtimeID}})
}
