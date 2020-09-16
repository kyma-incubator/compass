package spec

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

const specDefTable string = `"public"."specifications"`

var (
	tenantColumn     = "tenant_id"
	specColumns      = []string{"id", "tenant_id", "api_def_id", "event_def_id", "spec_data", "spec_format", "spec_type"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"spec_data", "spec_format", "spec_type"}
)

//go:generate mockery -name=SpecConverter -output=automock -outpkg=automock -case=underscore
type SpecConverter interface {
	FromEntity(entity Entity) model.Spec
	ToEntity(apiModel model.Spec) Entity
	APISpecInputFromSpec(spec *model.Spec, fr *model.FetchRequest) *model.APISpecInput
	EventSpecInputFromSpec(spec *model.Spec, fr *model.FetchRequest) *model.EventSpecInput
}

type pgRepository struct {
	singleGetter repo.SingleGetter
	creator      repo.Creator
	lister       repo.Lister
	updater      repo.Updater
	deleter      repo.Deleter
	existQuerier repo.ExistQuerier
	conv         SpecConverter
}

func NewRepository(conv SpecConverter) *pgRepository {
	return &pgRepository{
		singleGetter: repo.NewSingleGetter(resource.Spec, specDefTable, tenantColumn, specColumns),
		lister:       repo.NewLister(resource.Spec, specDefTable, tenantColumn, specColumns),
		creator:      repo.NewCreator(resource.Spec, specDefTable, specColumns),
		updater:      repo.NewUpdater(resource.Spec, specDefTable, updatableColumns, tenantColumn, idColumns),
		deleter:      repo.NewDeleter(resource.Spec, specDefTable, tenantColumn),
		existQuerier: repo.NewExistQuerier(resource.Spec, specDefTable, tenantColumn),
		conv:         conv,
	}
}

type SpecCollection []Entity

func (r SpecCollection) Len() int {
	return len(r)
}

func (r *pgRepository) ListForAPI(ctx context.Context, tenantID string, apiID string) ([]*model.Spec, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("api_def_id", apiID),
	}
	return r.list(ctx, tenantID, conditions)
}

func (r *pgRepository) ListForEvent(ctx context.Context, tenantID string, eventID string) ([]*model.Spec, error) {
	conditions := repo.Conditions{
		repo.NewEqualCondition("event_def_id", eventID),
	}
	return r.list(ctx, tenantID, conditions)
}

func (r *pgRepository) list(ctx context.Context, tenant string, conditions repo.Conditions) ([]*model.Spec, error) {
	var specCollection SpecCollection
	err := r.lister.List(ctx, tenant, &specCollection, conditions...)
	if err != nil {
		return nil, err
	}

	var items []*model.Spec

	for _, specEnt := range specCollection {
		m := r.conv.FromEntity(specEnt)

		items = append(items, &m)
	}

	return items, nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.Spec, error) {
	return r.GetByField(ctx, tenantID, "id", id)
}

func (r *pgRepository) GetByField(ctx context.Context, tenant, fieldName, fieldValue string) (*model.Spec, error) {
	var apiDefEntity Entity
	err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition(fieldName, fieldValue)}, repo.NoOrderBy, &apiDefEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting APIDefinition")
	}

	apiDefModel := r.conv.FromEntity(apiDefEntity)

	return &apiDefModel, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)
	err := r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

func (r *pgRepository) CreateMany(ctx context.Context, items []*model.Spec) error {
	for index, item := range items {
		entity := r.conv.ToEntity(*item)

		err := r.creator.Create(ctx, entity)
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}

	return nil
}

func (r *pgRepository) Update(ctx context.Context, item *model.Spec) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(*item)

	return r.updater.UpdateSingle(ctx, entity)
}

func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenantID, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}
