package runtime

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const runtimeTable string = `public.runtimes`

var (
	runtimeColumns = []string{"id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Runtime) (*Runtime, error)
	FromEntity(entity *Runtime) *model.Runtime
}

type pgRepository struct {
	existQuerier       repo.ExistQuerier
	ownerExistQuerier  repo.ExistQuerier
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	deleter            repo.Deleter
	pageableQuerier    repo.PageableQuerier
	lister             repo.Lister
	ownerLister        repo.Lister
	listerGlobal       repo.ListerGlobal
	creator            repo.Creator
	updater            repo.Updater
	conv               EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:       repo.NewExistQuerier(runtimeTable),
		ownerExistQuerier:  repo.NewExistQuerierWithOwnerCheck(runtimeTable),
		singleGetter:       repo.NewSingleGetter(runtimeTable, runtimeColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Runtime, runtimeTable, runtimeColumns),
		deleter:            repo.NewDeleter(runtimeTable),
		pageableQuerier:    repo.NewPageableQuerier(runtimeTable, runtimeColumns),
		lister:             repo.NewLister(runtimeTable, runtimeColumns),
		ownerLister:        repo.NewOwnerLister(runtimeTable, runtimeColumns, true),
		listerGlobal:       repo.NewListerGlobal(resource.Runtime, runtimeTable, runtimeColumns),
		creator:            repo.NewCreator(runtimeTable, runtimeColumns),
		updater:            repo.NewUpdater(runtimeTable, []string{"name", "description", "status_condition", "status_timestamp"}, []string{"id"}),
		conv:               conv,
	}
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Runtime, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// OwnerExists checks if runtime with given id and tenant exists and has owner access
func (r *pgRepository) OwnerExists(ctx context.Context, tenant, id string) (bool, error) {
	return r.ownerExistQuerier.Exists(ctx, resource.Runtime, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// ExistsByFiltersAndIDOwned checks if runtime with given id and filters in given tenant exists and has owner access
func (r *pgRepository) ExistsByFiltersAndIDOwned(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (bool, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return false, errors.Wrap(err, "while parsing tenant as UUID")
	}

	additionalConditions := repo.Conditions{repo.NewEqualCondition("id", id)}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return false, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	return r.ownerExistQuerier.Exists(ctx, resource.Runtime, tenant, additionalConditions)
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.Runtime, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error) {
	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, resource.Runtime, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := r.conv.FromEntity(&runtimeEnt)

	return runtimeModel, nil
}

// GetByFiltersAndID missing godoc
func (r *pgRepository) GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	additionalConditions := repo.Conditions{repo.NewEqualCondition("id", id)}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, resource.Runtime, tenant, additionalConditions, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := r.conv.FromEntity(&runtimeEnt)

	return runtimeModel, nil
}

// GetByFilters retrieves model.Runtime matching on the given label filters
func (r *pgRepository) GetByFilters(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	var runtimeEnt Runtime

	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	if err = r.singleGetter.Get(ctx, resource.Runtime, tenant, conditions, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&runtimeEnt)

	return appModel, nil
}

// GetByFiltersGlobal missing godoc
func (r *pgRepository) GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	filterSubquery, args, err := label.FilterQueryGlobal(model.RuntimeLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var additionalConditions repo.Conditions
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeEnt Runtime
	if err := r.singleGetterGlobal.GetGlobal(ctx, additionalConditions, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := r.conv.FromEntity(&runtimeEnt)

	return runtimeModel, nil
}

// ListByFiltersGlobal missing godoc
func (r *pgRepository) ListByFiltersGlobal(ctx context.Context, filters []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	filterSubquery, args, err := label.FilterQueryGlobal(model.RuntimeLabelableObject, label.IntersectSet, filters)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var additionalConditions repo.Conditions
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var entities RuntimeCollection
	if err := r.listerGlobal.ListGlobal(ctx, &entities, additionalConditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
}

// ListAll returns all runtimes in a tenant that match given label filter and owner check to false
func (r *pgRepository) ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	return r.listRuntimes(ctx, tenant, filter, false)
}

// ListOwnedRuntimes returns all runtimes in a tenant that match given label filter and owner check to true
func (r *pgRepository) ListOwnedRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	return r.listRuntimes(ctx, tenant, filter, true)
}

// RuntimeCollection missing godoc
type RuntimeCollection []Runtime

// Len missing godoc
func (r RuntimeCollection) Len() int {
	return len(r)
}

// List missing godoc
func (r *pgRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error) {
	var runtimesCollection RuntimeCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}
	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, resource.Runtime, tenant, pageSize, cursor, "name", &runtimesCollection, conditions...)

	if err != nil {
		return nil, err
	}

	items := r.multipleFromEntities(runtimesCollection)

	return &model.RuntimePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.Runtime) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	runtimeEnt, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}

	return r.creator.Create(ctx, resource.Runtime, tenant, runtimeEnt)
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.Runtime) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}
	runtimeEnt, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}
	return r.updater.UpdateSingle(ctx, resource.Runtime, tenant, runtimeEnt)
}

// GetOldestForFilters missing godoc
func (r *pgRepository) GetOldestForFilters(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	var additionalConditions repo.Conditions
	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	orderByParams := repo.OrderByParams{repo.NewAscOrderBy("creation_timestamp")}

	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, resource.Runtime, tenant, additionalConditions, orderByParams, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := r.conv.FromEntity(&runtimeEnt)

	return runtimeModel, nil
}

func (r *pgRepository) listRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, ownerCheck bool) ([]*model.Runtime, error) {
	var entities RuntimeCollection

	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	if ownerCheck {
		if err = r.ownerLister.List(ctx, resource.Runtime, tenant, &entities, conditions...); err != nil {
			return nil, err
		}
	} else {
		if err = r.lister.List(ctx, resource.Runtime, tenant, &entities, conditions...); err != nil {
			return nil, err
		}
	}

	return r.multipleFromEntities(entities), nil
}

func (r *pgRepository) multipleFromEntities(entities RuntimeCollection) []*model.Runtime {
	items := make([]*model.Runtime, 0, len(entities))
	for _, ent := range entities {
		model := r.conv.FromEntity(&ent)

		items = append(items, model)
	}
	return items
}
