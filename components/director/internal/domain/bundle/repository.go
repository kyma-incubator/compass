package mp_bundle

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const bundleTable string = `public.bundles`
const bundleInstanceAuthTable string = `public.bundle_instance_auths`
const bundleInstanceAuthBundleRefField string = `bundle_id`

var (
	bundleColumns    = []string{"id", "tenant_id", "app_id", "name", "description", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error"}
	tenantColumn     = "tenant_id"
	updatableColumns = []string{"name", "description", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error"}
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Bundle) (*Entity, error)
	FromEntity(entity *Entity) (*model.Bundle, error)
}

type pgRepository struct {
	existQuerier    repo.ExistQuerier
	singleGetter    repo.SingleGetter
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator
	updater         repo.Updater
	conv            EntityConverter
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:    repo.NewExistQuerier(resource.Bundle, bundleTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		deleter:         repo.NewDeleter(resource.Bundle, bundleTable, tenantColumn),
		pageableQuerier: repo.NewPageableQuerier(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		creator:         repo.NewCreator(resource.Bundle, bundleTable, bundleColumns),
		updater:         repo.NewUpdater(resource.Bundle, bundleTable, updatableColumns, tenantColumn, []string{"id"}),
		conv:            conv,
	}
}

type BundleCollection []Entity

func (r BundleCollection) Len() int {
	return len(r)
}

func (r *pgRepository) Create(ctx context.Context, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bndlEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	log.C(ctx).Debugf("Persisting Bundle entity with id %s to db", model.ID)
	return r.creator.Create(ctx, bndlEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bndlEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	return r.updater.UpdateSingle(ctx, bndlEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error) {
	var bndlEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &bndlEnt); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(&bndlEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Bundle from Entity")
	}

	return bndlModel, nil
}

func (r *pgRepository) GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("app_id", applicationID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return bndlModel, nil
}

func (r *pgRepository) GetByInstanceAuthID(ctx context.Context, tenant string, instanceAuthID string) (*model.Bundle, error) {
	var bndlEnt Entity

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	prefixedFieldNames := str.PrefixStrings(bundleColumns, "b.")
	stmt := fmt.Sprintf(`SELECT %s FROM %s AS b JOIN %s AS a on a.%s=b.id where a.tenant_id=$1 AND a.id=$2`,
		strings.Join(prefixedFieldNames, ", "),
		bundleTable,
		bundleInstanceAuthTable,
		bundleInstanceAuthBundleRefField)

	err = persist.Get(&bndlEnt, stmt, tenant, instanceAuthID)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "while getting Bundle by Instance Auth ID")
	}

	bndlModel, err := r.conv.FromEntity(&bndlEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return bndlModel, nil
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.BundlePage, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("app_id", applicationID),
	}

	var bundleCollection BundleCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &bundleCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.Bundle

	for _, bndlEnt := range bundleCollection {
		m, err := r.conv.FromEntity(&bndlEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Bundle model from entity")
		}
		items = append(items, m)
	}

	return &model.BundlePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}
