package formationassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.formation_assignments`

var (
	idTableColumns        = []string{"id"}
	updatableTableColumns = []string{"state", "value"}
	tableColumns          = []string{"id", "formation_id", "tenant_id", "source", "source_type", "target", "target_type", "state", "value"}
	tenantColumn          = "tenant_id"
)

// EntityConverter converts between the internal model and entity
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.FormationAssignment) *Entity
	FromEntity(entity *Entity) *model.FormationAssignment
}

type repository struct {
	creator               repo.CreatorGlobal
	getter                repo.SingleGetter
	globalGetter          repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerier
	unionLister           repo.UnionLister
	lister                repo.Lister
	conditionLister       repo.ConditionTreeLister
	updaterGlobal         repo.UpdaterGlobal
	deleter               repo.Deleter
	existQuerier          repo.ExistQuerier
	conv                  EntityConverter
}

// NewRepository creates a new FormationAssignment repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:               repo.NewCreatorGlobal(resource.FormationAssignment, tableName, tableColumns),
		getter:                repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		globalGetter:          repo.NewSingleGetterGlobal(resource.FormationAssignment, tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		unionLister:           repo.NewUnionListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		lister:                repo.NewListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		conditionLister:       repo.NewConditionTreeListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		updaterGlobal:         repo.NewUpdaterWithEmbeddedTenant(resource.FormationAssignment, tableName, updatableTableColumns, tenantColumn, idTableColumns),
		deleter:               repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		existQuerier:          repo.NewExistQuerierWithEmbeddedTenant(tableName, tenantColumn),
		conv:                  conv,
	}
}

// Create creates a new Formation Assignment in the database with the fields from the model
func (r *repository) Create(ctx context.Context, item *model.FormationAssignment) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Persisting Formation Assignment entity with ID: %q", item.ID)
	return r.creator.Create(ctx, r.conv.ToEntity(item))
}

// GetByTargetAndSource queries for a single Formation Assignment matching by a given Target and Source
func (r *repository) GetByTargetAndSource(ctx context.Context, target, source, tenantID string) (*model.FormationAssignment, error) {
	var entity Entity
	if err := r.getter.Get(ctx, resource.FormationAssignment, tenantID, repo.Conditions{repo.NewEqualCondition("target", target), repo.NewEqualCondition("source", source)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// Get queries for a single Formation Assignment matching by a given ID
func (r *repository) Get(ctx context.Context, id, tenantID string) (*model.FormationAssignment, error) {
	var entity Entity
	if err := r.getter.Get(ctx, resource.FormationAssignment, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// GetGlobalByID retrieves formation assignment matching ID `id` globally without tenant parameter
func (r *repository) GetGlobalByID(ctx context.Context, id string) (*model.FormationAssignment, error) {
	var entity Entity
	if err := r.globalGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// GetGlobalByIDAndFormationID retrieves formation assignment matching ID `id` and formation ID `formationID` globally, without tenant parameter
func (r *repository) GetGlobalByIDAndFormationID(ctx context.Context, id, formationID string) (*model.FormationAssignment, error) {
	var entity Entity
	if err := r.globalGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id), repo.NewEqualCondition("formation_id", formationID)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// GetForFormation retrieves Formation Assignment with the provided `id` associated to Formation with id `formationID` from the database if it exists
func (r *repository) GetForFormation(ctx context.Context, tenantID, id, formationID string) (*model.FormationAssignment, error) {
	var formationAssignmentEnt Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("formation_id", formationID),
	}

	if err := r.getter.Get(ctx, resource.FormationAssignment, tenantID, conditions, repo.NoOrderBy, &formationAssignmentEnt); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&formationAssignmentEnt), nil
}

// GetBySourceAndTarget retrieves formation assignment by source and target
func (r *repository) GetBySourceAndTarget(ctx context.Context, tenantID, formationID, sourceID, targetID string) (*model.FormationAssignment, error) {
	var formationAssignmentEnt Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("formation_id", formationID),
		repo.NewEqualCondition("source", sourceID),
		repo.NewEqualCondition("target", targetID),
	}

	if err := r.getter.Get(ctx, resource.FormationAssignment, tenantID, conditions, repo.NoOrderBy, &formationAssignmentEnt); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&formationAssignmentEnt), nil
}

// GetReverseBySourceAndTarget retrieves reverse formation assignment by source and target
func (r *repository) GetReverseBySourceAndTarget(ctx context.Context, tenantID, formationID, sourceID, targetID string) (*model.FormationAssignment, error) {
	return r.GetBySourceAndTarget(ctx, tenantID, formationID, targetID, sourceID)
}

// List queries for all Formation Assignment sorted by ID and paginated by the pageSize and cursor parameters
func (r *repository) List(ctx context.Context, pageSize int, cursor, tenantID string) (*model.FormationAssignmentPage, error) {
	var entityCollection EntityCollection
	page, totalCount, err := r.pageableQuerierGlobal.List(ctx, resource.FormationAssignment, tenantID, pageSize, cursor, "id", &entityCollection)
	if err != nil {
		return nil, err
	}

	return &model.FormationAssignmentPage{
		Data:       r.multipleFromEntities(entityCollection),
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListByFormationIDs retrieves a page of Formation Assignment objects for each formationID from the database that are visible for `tenantID`
func (r *repository) ListByFormationIDs(ctx context.Context, tenantID string, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error) {
	var formationAssignmentCollection EntityCollection

	orderByColumns := repo.OrderByParams{repo.NewAscOrderBy("formation_id"), repo.NewAscOrderBy("id")}

	counts, err := r.unionLister.List(ctx, resource.FormationAssignment, tenantID, formationIDs, "formation_id", pageSize, cursor, orderByColumns, &formationAssignmentCollection)
	if err != nil {
		return nil, err
	}

	formationAssignmentByFormationID := map[string][]*model.FormationAssignment{}
	for _, faEntity := range formationAssignmentCollection {
		formationAssignmentByFormationID[faEntity.FormationID] = append(formationAssignmentByFormationID[faEntity.FormationID], r.conv.FromEntity(faEntity))
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	faPages := make([]*model.FormationAssignmentPage, 0, len(formationIDs))
	for _, formationID := range formationIDs {
		totalCount := counts[formationID]
		hasNextPage := false
		endCursor := ""
		if totalCount > offset+len(formationAssignmentByFormationID[formationID]) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		faPages = append(faPages, &model.FormationAssignmentPage{Data: formationAssignmentByFormationID[formationID], TotalCount: totalCount, PageInfo: page})
	}

	return faPages, nil
}

// ListByFormationIDsNoPaging retrieves all Formation Assignment objects for each formationID from the database that are visible for `tenantID`
func (r *repository) ListByFormationIDsNoPaging(ctx context.Context, tenantID string, formationIDs []string) ([][]*model.FormationAssignment, error) {
	if len(formationIDs) == 0 {
		return nil, nil
	}

	var formationAssignmentCollection EntityCollection

	conditions := repo.NewInConditionForStringValues("formation_id", formationIDs)

	if err := r.lister.List(ctx, resource.FormationAssignment, tenantID, &formationAssignmentCollection, conditions); err != nil {
		return nil, err
	}

	formationAssignmentByFormationID := map[string][]*model.FormationAssignment{}
	for _, faEntity := range formationAssignmentCollection {
		formationAssignmentByFormationID[faEntity.FormationID] = append(formationAssignmentByFormationID[faEntity.FormationID], r.conv.FromEntity(faEntity))
	}

	formationAssignmentsPerFormation := make([][]*model.FormationAssignment, 0, len(formationIDs))
	for _, formationID := range formationIDs {
		formationAssignmentsPerFormation = append(formationAssignmentsPerFormation, formationAssignmentByFormationID[formationID])
	}

	return formationAssignmentsPerFormation, nil
}

// ListAllForObject retrieves all FormationAssignment objects for formation with ID `formationID` that have objectID as `target` or `source` from the database that are visible for `tenant`
func (r *repository) ListAllForObject(ctx context.Context, tenant, formationID, objectID string) ([]*model.FormationAssignment, error) {
	var entities EntityCollection
	conditions := repo.And(
		&repo.ConditionTree{Operand: repo.NewEqualCondition("formation_id", formationID)},
		repo.Or(repo.ConditionTreesFromConditions([]repo.Condition{
			repo.NewEqualCondition("source", objectID),
			repo.NewEqualCondition("target", objectID),
		})...))

	if err := r.conditionLister.ListConditionTree(ctx, resource.FormationAssignment, tenant, &entities, conditions); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
}

// ListAllForObjectIDs retrieves all FormationAssignment objects for formation with ID `formationID` that have any of the objectIDs as `target` or `source` from the database that are visible for `tenant`
func (r *repository) ListAllForObjectIDs(ctx context.Context, tenant, formationID string, objectIDs []string) ([]*model.FormationAssignment, error) {
	var entities EntityCollection
	conditions := repo.And(
		&repo.ConditionTree{Operand: repo.NewEqualCondition("formation_id", formationID)},
		repo.Or(repo.ConditionTreesFromConditions([]repo.Condition{
			repo.NewInConditionForStringValues("source", objectIDs),
			repo.NewInConditionForStringValues("target", objectIDs),
		})...))

	if err := r.conditionLister.ListConditionTree(ctx, resource.FormationAssignment, tenant, &entities, conditions); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
}

// ListForIDs missing godoc
func (r *repository) ListForIDs(ctx context.Context, tenant string, ids []string) ([]*model.FormationAssignment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var entitiesWithIDs EntityCollection
	conditions := repo.NewInConditionForStringValues("id", ids)

	if err := r.lister.List(ctx, resource.FormationAssignment, tenant, &entitiesWithIDs, conditions); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entitiesWithIDs), nil
}

// Update updates the Formation Assignment matching the ID of the input model
func (r *repository) Update(ctx context.Context, model *model.FormationAssignment) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// Delete deletes a Formation Assignment with given ID
func (r *repository) Delete(ctx context.Context, id, tenantID string) error {
	return r.deleter.DeleteOne(ctx, resource.FormationAssignment, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists check if a Formation Assignment with given ID exists
func (r *repository) Exists(ctx context.Context, id, tenantID string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.FormationAssignment, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) multipleFromEntities(entities EntityCollection) []*model.FormationAssignment {
	items := make([]*model.FormationAssignment, 0, len(entities))
	for _, ent := range entities {
		items = append(items, r.conv.FromEntity(ent))
	}
	return items
}
