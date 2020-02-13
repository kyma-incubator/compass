package mp_package

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const packageTable string = `public.packages`

var (
	packageColumns = []string{"id", "tenant_id", "app_id", "name", "description", "instance_auth_request_json_schema", "default_instance_auth"}
	tenantColumn   = "tenant_id"
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Package) (*Entity, error)
	FromEntity(entity *Entity) (*model.Package, error)
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
		existQuerier:    repo.NewExistQuerier(packageTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(packageTable, tenantColumn, packageColumns),
		deleter:         repo.NewDeleter(packageTable, tenantColumn),
		pageableQuerier: repo.NewPageableQuerier(packageTable, tenantColumn, packageColumns),
		creator:         repo.NewCreator(packageTable, packageColumns),
		updater:         repo.NewUpdater(packageTable, []string{"name", "description", "instance_auth_request_json_schema", "default_instance_auth"}, tenantColumn, []string{"id"}),
		conv:            conv,
	}
}

type PackageCollection []Entity

func (r PackageCollection) Len() int {
	return len(r)
}

func (r *pgRepository) Create(ctx context.Context, model *model.Package) error {
	if model == nil {
		return errors.New("model can not be empty")
	}

	pkgEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Package entity")
	}

	return r.creator.Create(ctx, pkgEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Package) error {
	if model == nil {
		return errors.New("model can not be empty")
	}

	pkgEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Package entity")
	}

	return r.updater.UpdateSingle(ctx, pkgEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Package, error) {
	var pkgEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &pkgEnt); err != nil {
		return nil, err
	}

	pkgModel, err := r.conv.FromEntity(&pkgEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Package from Entity")
	}

	return pkgModel, nil
}

func (r *pgRepository) GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Package, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("app_id", applicationID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	pkgModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Package model from entity")
	}

	return pkgModel, nil
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.PackagePage, error) {
	pkgCond := fmt.Sprintf("%s = '%s'", "app_id", applicationID)
	var packageCollection PackageCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &packageCollection, pkgCond)
	if err != nil {
		return nil, err
	}

	var items []*model.Package

	for _, pkgEnt := range packageCollection {
		m, err := r.conv.FromEntity(&pkgEnt)
		if err != nil {
			return nil, errors.Wrap(err, "while creating Package model from entity")
		}
		items = append(items, m)
	}

	return &model.PackagePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}
