package mp_package

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const bundleTable string = `public.bundles`

var (
	bundleColumns = []string{"id", "tenant_id", "app_id", "title", "short_description", "description", "instance_auth_request_json_schema", "default_instance_auth", "tags", "last_updated", "extensions"}
	tenantColumn  = "tenant_id"
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
		updater:         repo.NewUpdater(resource.Bundle, bundleTable, []string{"name", "description", "instance_auth_request_json_schema", "default_instance_auth"}, tenantColumn, []string{"id"}),
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

	bundleEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	return r.creator.Create(ctx, bundleEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Bundle) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	bundleEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Bundle entity")
	}

	return r.updater.UpdateSingle(ctx, bundleEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error) {
	var bundleEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &bundleEnt); err != nil {
		return nil, err
	}

	bundleModel, err := r.conv.FromEntity(&bundleEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Bundle from Entity")
	}

	return bundleModel, nil
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

	bundleModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return bundleModel, nil
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

	for _, bundleEnt := range bundleCollection {
		m, err := r.conv.FromEntity(&bundleEnt)
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
