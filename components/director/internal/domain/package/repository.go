package ordpackage

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const packageTable string = `public.packages`

var (
	tenantColumn   = "tenant_id" // TODO: <storage-redesign> delete this column
	packageColumns = []string{"id", "app_id", "ord_id", "vendor", "title", "short_description",
		"description", "version", "package_links", "links", "licence_type", "tags", "countries", "labels", "policy_level",
		"custom_policy_level", "part_of_products", "line_of_business", "industry", "resource_hash"}
	updatableColumns = []string{"vendor", "title", "short_description", "description", "version", "package_links", "links",
		"licence_type", "tags", "countries", "labels", "policy_level", "custom_policy_level", "part_of_products", "line_of_business", "industry", "resource_hash"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	ToEntity(in *model.Package) *Entity
	FromEntity(entity *Entity) (*model.Package, error)
}

type pgRepository struct {
	conv         EntityConverter
	existQuerier repo.ExistQuerier
	lister       repo.Lister
	singleGetter repo.SingleGetter
	deleter      repo.Deleter
	creator      repo.Creator
	updater      repo.Updater
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:         conv,
		existQuerier: repo.NewExistQuerier(resource.Package, packageTable, tenantColumn),
		lister:       repo.NewLister(resource.Package, packageTable, tenantColumn, packageColumns),
		singleGetter: repo.NewSingleGetter(resource.Package, packageTable, tenantColumn, packageColumns),
		deleter:      repo.NewDeleter(resource.Package, packageTable, tenantColumn),
		creator:      repo.NewCreator(resource.Package, packageTable, packageColumns),
		updater:      repo.NewUpdater(resource.Package, packageTable, updatableColumns, tenantColumn, []string{"id"}),
	}
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Package entity with id %q", model.ID)
	return r.creator.Create(ctx, tenant, r.conv.ToEntity(model))
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Package entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, r.conv.ToEntity(model))
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Package entity with id %q", id)
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Package, error) {
	log.C(ctx).Debugf("Getting Package entity with id %q", id)
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

// ListByApplicationID missing godoc
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Package, error) {
	pkgCollection := pkgCollection{}
	if err := r.lister.List(ctx, tenantID, &pkgCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	pkgs := make([]*model.Package, 0, pkgCollection.Len())
	for _, pkg := range pkgCollection {
		pkgModel, err := r.conv.FromEntity(&pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkgModel)
	}
	return pkgs, nil
}

type pkgCollection []Entity

// Len missing godoc
func (pc pkgCollection) Len() int {
	return len(pc)
}
