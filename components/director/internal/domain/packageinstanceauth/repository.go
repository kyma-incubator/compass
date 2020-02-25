package packageinstanceauth

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.package_instance_auths`

var (
	tenantColumn     = "tenant_id"
	idColumns        = []string{"id"}
	updatableColumns = []string{"auth_value", "status_condition", "status_timestamp", "status_message", "status_reason"}
	tableColumns     = []string{"id", "tenant_id", "package_id", "context", "input_params", "auth_value", "status_condition", "status_timestamp", "status_message", "status_reason"}
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in model.PackageInstanceAuth) (Entity, error)
	FromEntity(entity Entity) (model.PackageInstanceAuth, error)
}

type repository struct {
	creator      repo.Creator
	singleGetter repo.SingleGetter
	lister       repo.Lister
	updater      repo.Updater
	deleter      repo.Deleter
	conv         EntityConverter
}

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:      repo.NewCreator(tableName, tableColumns),
		singleGetter: repo.NewSingleGetter(tableName, tenantColumn, tableColumns),
		lister:       repo.NewLister(tableName, tenantColumn, tableColumns),
		deleter:      repo.NewDeleter(tableName, tenantColumn),
		updater:      repo.NewUpdater(tableName, updatableColumns, tenantColumn, idColumns),
		conv:         conv,
	}
}

func (r *repository) Create(ctx context.Context, item *model.PackageInstanceAuth) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting PackageInstanceAuth model to entity")
	}

	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, tenantID string, id string) (*model.PackageInstanceAuth, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting PackageInstanceAuth entity to model")
	}

	return &itemModel, nil
}

func (r *repository) GetForPackage(ctx context.Context, tenant string, id string, packageID string) (*model.PackageInstanceAuth, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("package_id", packageID),
	}
	if err := r.singleGetter.Get(ctx, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	pkgModel, err := r.conv.FromEntity(ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Package model from entity")
	}

	return &pkgModel, nil
}

func (r *repository) ListByPackageID(ctx context.Context, tenantID string, packageID string) ([]*model.PackageInstanceAuth, error) {
	var entities Collection

	err := r.lister.List(ctx, tenantID, &entities, fmt.Sprintf("package_id = %s", pq.QuoteLiteral(packageID)))

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

func (r *repository) Update(ctx context.Context, item *model.PackageInstanceAuth) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.updater.UpdateSingle(ctx, entity)
}

func (r *repository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) multipleFromEntities(entities Collection) ([]*model.PackageInstanceAuth, error) {
	var items []*model.PackageInstanceAuth
	for _, ent := range entities {
		m, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating PackageInstanceAuth model from entity")
		}
		items = append(items, &m)
	}
	return items, nil
}
