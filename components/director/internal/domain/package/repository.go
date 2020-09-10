package mp_package

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const packageTable string = `public.packages`
const packageBundleTable string = `public.package_bundles`

var (
	packageColumns = []string{"id", "od_id", "tenant_id", "app_id", "title", "short_description", "description", "version",
		"licence", "licence_type", "terms_of_service", "logo", "image", "provider", "tags", "last_updated", "extensions"}
	updatableColumns = []string{"title", "short_description", "description", "version",
		"licence", "licence_type", "terms_of_service", "logo", "image", "provider", "tags", "last_updated", "extensions"}
	packageBundleColumns = []string{"package_id", "bundle_id"}
	tenantColumn         = "tenant_id"
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Package) (*Entity, error)
	FromEntity(entity *Entity) (*model.Package, error)
}

type pgRepository struct {
	existQuerier             repo.ExistQuerier
	relationshipExistQuerier repo.ExistQuerierGlobal
	singleGetter             repo.SingleGetter
	deleter                  repo.Deleter
	pageableQuerier          repo.PageableQuerier
	creator                  repo.Creator
	relationshipCreator      repo.Creator
	updater                  repo.Updater
	conv                     EntityConverter
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:             repo.NewExistQuerier(resource.Package, packageTable, tenantColumn),
		relationshipExistQuerier: repo.NewExistQuerierGlobal(resource.PackageBundle, packageBundleTable),
		singleGetter:             repo.NewSingleGetter(resource.Package, packageTable, tenantColumn, packageColumns),
		deleter:                  repo.NewDeleter(resource.Package, packageTable, tenantColumn),
		pageableQuerier:          repo.NewPageableQuerier(resource.Package, packageTable, tenantColumn, packageColumns),
		creator:                  repo.NewCreator(resource.Package, packageTable, packageColumns),
		relationshipCreator:      repo.NewCreator(resource.PackageBundle, packageBundleTable, packageBundleColumns),
		updater:                  repo.NewUpdater(resource.Package, packageTable, updatableColumns, tenantColumn, []string{"id"}),
		conv:                     conv,
	}
}

type PackageCollection []Entity

func (r PackageCollection) Len() int {
	return len(r)
}

func (r *pgRepository) Create(ctx context.Context, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	packageEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Package entity")
	}

	return r.creator.Create(ctx, packageEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Package) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	packageEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Package entity")
	}

	return r.updater.UpdateSingle(ctx, packageEnt)
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) ExistsByCondition(ctx context.Context, tenant string, conds repo.Conditions) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, conds)
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Package, error) {
	return r.GetByField(ctx, tenant, "id", id)
}

func (r *pgRepository) GetByField(ctx context.Context, tenant,fieldName, fieldValue string) (*model.Package, error) {
	var packageEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition(fieldName, fieldValue)}, repo.NoOrderBy, &packageEnt); err != nil {
		return nil, err
	}

	packageModel, err := r.conv.FromEntity(&packageEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Package from Entity")
	}

	return packageModel, nil
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

	packageModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Package model from entity")
	}

	return packageModel, nil
}

func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.PackagePage, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("app_id", applicationID),
	}

	var packageCollection PackageCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, "id", &packageCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.Package

	for _, packageEnt := range packageCollection {
		m, err := r.conv.FromEntity(&packageEnt)
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

func (r *pgRepository) AssociateBundle(ctx context.Context, id, bundleID string) error {
	if len(id) == 0 || len(bundleID) == 0 {
		return apperrors.NewInternalError("id or bundleID can not be empty")
	}
	entity := struct {
		PackageID string `db:"package_id"`
		BundleID  string `db:"bundle_id"`
	}{
		PackageID: id,
		BundleID:  bundleID,
	}

	exists, err := r.relationshipExistQuerier.ExistsGlobal(ctx, []repo.Condition{
		repo.NewEqualCondition("package_id", id),
		repo.NewEqualCondition("bundle_id", bundleID),
	})
	if err != nil {
		return err
	}
	if !exists {
		return r.relationshipCreator.Create(ctx, entity)
	}
	return nil
}
