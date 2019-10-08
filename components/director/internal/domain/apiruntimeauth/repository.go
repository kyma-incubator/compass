package apiruntimeauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"
)

const tableName string = `public.api_runtime_auths`

var (
	tableColumns = []string{"id", "tenant_id", "runtime_id", "api_def_id", "value"}
	tenantColumn = "tenant_id"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.APIRuntimeAuth) (Entity, error)
	FromEntity(in Entity) (model.APIRuntimeAuth, error)
}

type pgRepository struct {
	singleGetter repo.SingleGetter
	lister       repo.Lister
	upserter     repo.Upserter
	deleter      repo.Deleter

	conv Converter
}

func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		singleGetter: repo.NewSingleGetter(tableName, tenantColumn, tableColumns),
		lister:       repo.NewLister(tableName, tenantColumn, tableColumns),
		upserter:     repo.NewUpserter(tableName, tableColumns, []string{"tenant_id", "runtime_id", "api_def_id"}, []string{"value"}),
		deleter:      repo.NewDeleter(tableName, tenantColumn),
		conv:         conv,
	}
}

func (r *pgRepository) Get(ctx context.Context, tenant string, apiID string, runtimeID string) (*model.APIRuntimeAuth, error) {
	var ent Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{{Field: "runtime_id", Val: runtimeID}, {Field: "api_def_id", Val: apiID}}, &ent); err != nil {
		return nil, err
	}

	apiRtmAuthModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating api runtime auth model from entity")
	}

	return &apiRtmAuthModel, nil
}

func (r *pgRepository) GetOrDefault(ctx context.Context, tenant string, apiID string, runtimeID string) (*model.APIRuntimeAuth, error) {
	var ent Entity

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ara.id, $2 AS api_def_id,
	COALESCE(ara.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value
	FROM (SELECT * FROM runtimes WHERE id = $3) AS r
	LEFT OUTER JOIN (SELECT * FROM api_runtime_auths
	WHERE api_def_id = $2 AND runtime_id = $3 AND tenant_id = $1) AS ara ON ara.runtime_id = r.id`

	err = persist.Get(&ent, stmt, tenant, apiID, runtimeID)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "while getting api runtime auth or default api runtime auth from DB")
	}

	apiRtmAuthModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating api runtime auth model from entity")
	}

	return &apiRtmAuthModel, nil
}

type APIAPIRtmAuthCollection []Entity

func (r APIAPIRtmAuthCollection) Len() int {
	return len(r)
}

func (r *pgRepository) ListForAllRuntimes(ctx context.Context, tenant string, apiID string) ([]model.APIRuntimeAuth, error) {
	var apiRtmAuthsCollection APIAPIRtmAuthCollection

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ara.id, $2 AS api_def_id,
			coalesce(ara.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value
    		FROM (SELECT * FROM api_runtime_auths WHERE api_def_id = $2 AND tenant_id = $1) AS ara
    		RIGHT OUTER JOIN runtimes AS r ON ara.runtime_id = r.id WHERE r.tenant_id = $1`

	err = persist.Select(&apiRtmAuthsCollection, stmt, tenant, apiID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching list of objects from DB")
	}

	var items []model.APIRuntimeAuth

	for _, ent := range apiRtmAuthsCollection {
		m, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating api runtime auth model from entity")
		}

		items = append(items, m)
	}

	return items, nil
}

func (r *pgRepository) Upsert(ctx context.Context, item model.APIRuntimeAuth) error {
	runtimeEnt, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating api runtime auth entity from model")
	}

	return r.upserter.Upsert(ctx, runtimeEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant string, apiID string, runtimeID string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "api_def_id", Val: apiID}, {Field: "runtime_id", Val: runtimeID}})
}
