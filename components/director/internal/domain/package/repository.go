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

const (
	packageTable               = `public.packages`
	appTemplateVersionIDColumn = "app_template_version_id"
	appIDColumn                = "app_id"
)

var (
	packageColumns = []string{"id", "app_id", "app_template_version_id", "ord_id", "vendor", "title", "short_description",
		"description", "version", "package_links", "links", "licence_type", "tags", "countries", "labels", "policy_level",
		"custom_policy_level", "part_of_products", "line_of_business", "industry", "resource_hash", "documentation_labels", "support_info"}
	updatableColumns = []string{"vendor", "title", "short_description", "description", "version", "package_links", "links",
		"licence_type", "tags", "countries", "labels", "policy_level", "custom_policy_level", "part_of_products", "line_of_business", "industry", "resource_hash", "documentation_labels", "support_info"}
)

// EntityConverter missing godoc
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Package) *Entity
	FromEntity(entity *Entity) (*model.Package, error)
}

type pgRepository struct {
	conv               EntityConverter
	existQuerier       repo.ExistQuerier
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal
	creator            repo.Creator
	creatorGlobal      repo.CreatorGlobal
	updater            repo.Updater
	updaterGlobal      repo.UpdaterGlobal
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:               conv,
		existQuerier:       repo.NewExistQuerier(packageTable),
		lister:             repo.NewLister(packageTable, packageColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Package, packageTable, packageColumns),
		singleGetter:       repo.NewSingleGetter(packageTable, packageColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Package, packageTable, packageColumns),
		deleter:            repo.NewDeleter(packageTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Package, packageTable),
		creator:            repo.NewCreator(packageTable, packageColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.Package, packageTable, packageColumns),
		updater:            repo.NewUpdater(packageTable, updatableColumns, []string{"id"}),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Package, packageTable, updatableColumns, []string{"id"}),
	}
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Package entity with id %q", model.ID)
	return r.creator.Create(ctx, resource.Package, tenant, r.conv.ToEntity(model))
}

// CreateGlobal creates a packages globally without tenant isolation
func (r *pgRepository) CreateGlobal(ctx context.Context, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Package entity with id %q", model.ID)
	return r.creatorGlobal.Create(ctx, r.conv.ToEntity(model))
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Package entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, resource.Package, tenant, r.conv.ToEntity(model))
}

// UpdateGlobal updates a package globally without tenant isolation
func (r *pgRepository) UpdateGlobal(ctx context.Context, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Package entity with id %q", model.ID)
	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Package entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.Package, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal deletes a Package without tenant isolation
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting Package entity with id %q", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Package, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Package, error) {
	log.C(ctx).Debugf("Getting Package entity with id %q", id)
	var pkgEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Package, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &pkgEnt); err != nil {
		return nil, err
	}

	pkgModel, err := r.conv.FromEntity(&pkgEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Package from Entity")
	}

	return pkgModel, nil
}

// GetByIDGlobal gets a package by ID without tenant isolation
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.Package, error) {
	log.C(ctx).Debugf("Getting Package entity with id %q", id)
	var pkgEnt Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &pkgEnt); err != nil {
		return nil, err
	}

	pkgModel, err := r.conv.FromEntity(&pkgEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Package from Entity")
	}

	return pkgModel, nil
}

// ListByResourceID lists Packages by a given resource type and resource ID
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.Package, error) {
	pkgCollection := pkgCollection{}

	var condition repo.Condition
	var err error
	if resourceType == resource.Application {
		condition = repo.NewEqualCondition(appIDColumn, resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.Package, tenantID, &pkgCollection, condition)
	} else {
		condition = repo.NewEqualCondition(appTemplateVersionIDColumn, resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &pkgCollection, condition)
	}
	if err != nil {
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
