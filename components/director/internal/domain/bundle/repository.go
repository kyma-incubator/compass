package mp_bundle

import (
	"context"
	"fmt"
	"strings"

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
const packageBundleTable string = `public.package_bundles`

var (
	bundleColumns        = []string{"id", "od_id", "tenant_id", "app_id", "title", "short_description", "description", "instance_auth_request_json_schema", "default_instance_auth", "tags", "last_updated", "extensions"}
	packageBundleColumns = []string{"package_id", "bundle_id"}
	tenantColumn         = "tenant_id"
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
	listerGlobal    repo.ListerGlobal
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
		listerGlobal:    repo.NewListerGlobal(resource.PackageBundle, packageBundleTable, packageBundleColumns),
		pageableQuerier: repo.NewPageableQuerier(resource.Bundle, bundleTable, tenantColumn, bundleColumns),
		creator:         repo.NewCreator(resource.Bundle, bundleTable, bundleColumns),
		updater:         repo.NewUpdater(resource.Bundle, bundleTable, []string{"title", "short_description", "description", "instance_auth_request_json_schema", "default_instance_auth", "tags", "last_updated", "extensions"}, tenantColumn, []string{"id"}),
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

func (r *pgRepository) ExistsByCondition(ctx context.Context, tenant string, conds repo.Conditions) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, conds)
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error) {
	return r.GetByConditions(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByConditions(ctx context.Context, tenant string, conds repo.Conditions) (*model.Bundle, error) {
	var bundleEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, conds, repo.NoOrderBy, &bundleEnt); err != nil {
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

func (r *pgRepository) GetByInstanceAuthID(ctx context.Context, tenant string, instanceAuthID string) (*model.Bundle, error) {
	var bundleEnt Entity

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	prefixedFieldNames := str.PrefixStrings(bundleColumns, "p.")
	stmt := fmt.Sprintf(`SELECT %s FROM %s AS p JOIN %s AS a on a.%s=p.id where a.tenant_id=$1 AND a.id=$2`,
		strings.Join(prefixedFieldNames, ", "),
		bundleTable,
		bundleInstanceAuthTable,
		bundleInstanceAuthBundleRefField)

	err = persist.Get(&bundleEnt, stmt, tenant, instanceAuthID)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "while getting Bundle by Instance Auth ID")
	}

	bundleModel, err := r.conv.FromEntity(&bundleEnt)
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

func (r *pgRepository) GetForPackage(ctx context.Context, tenantID, id string, _ string) (*model.Bundle, error) {
	return r.GetByID(ctx, tenantID, id)
}

func (r *pgRepository) ListByPackageID(ctx context.Context, tenantID, packageID string, pageSize int, cursor string) (*model.BundlePage, error) {
	bundleIDs, err := r.getBundleIDsByPackageID(ctx, packageID)
	if err != nil {
		return nil, err
	}

	conditions := repo.Conditions{
		repo.NewInConditionForStringValues("id", bundleIDs),
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

func (r *pgRepository) getBundleIDsByPackageID(ctx context.Context, packageID string) ([]string, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("package_id", packageID),
	}

	var packageBundleCollection PackageBundleCollection

	err := r.listerGlobal.ListGlobal(ctx, &packageBundleCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var bundleIDs []string
	for _, entity := range packageBundleCollection {
		bundleIDs = append(bundleIDs, entity.BundleID)
	}

	return bundleIDs, nil
}

type PackageBundle struct {
	PackageID string `db:"package_id"`
	BundleID  string `db:"bundle_id"`
}

type PackageBundleCollection []PackageBundle

func (rpc PackageBundleCollection) Len() int {
	return len(rpc)
}
