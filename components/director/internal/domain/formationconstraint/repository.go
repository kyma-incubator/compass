package formationconstraint

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.formation_constraints`
const idColumn string = "id"

var (
	tableColumns = []string{"id", "name", "constraint_type", "target_operation", "operator", "resource_type", "resource_subtype", "operator_scope", "input_template", "constraint_scope"}
)

// EntityConverter converts between the internal model and entity
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.FormationConstraint) *Entity
	FromEntity(entity *Entity) *model.FormationConstraint
	MultipleFromEntity(in EntityCollection) []*model.FormationConstraint
}

type repository struct {
	conditionTreeLister repo.ConditionTreeListerGlobal
	lister              repo.ListerGlobal
	creator             repo.CreatorGlobal
	singleGetter        repo.SingleGetterGlobal
	deleter             repo.DeleterGlobal
	conv                EntityConverter
}

// NewRepository creates a new FormationConstraint repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		conditionTreeLister: repo.NewConditionTreeListerGlobal(tableName, tableColumns),
		lister:              repo.NewListerGlobal(resource.FormationConstraint, tableName, tableColumns),
		creator:             repo.NewCreatorGlobal(resource.FormationConstraint, tableName, tableColumns),
		singleGetter:        repo.NewSingleGetterGlobal(resource.FormationConstraint, tableName, tableColumns),
		deleter:             repo.NewDeleterGlobal(resource.FormationConstraint, tableName),
		conv:                conv,
	}
}

// Create stores new record in the database
func (r *repository) Create(ctx context.Context, item *model.FormationConstraint) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Formation Constraint with id %s to entity", item.ID)
	entity := r.conv.ToEntity(item)

	log.C(ctx).Debugf("Persisting Formation Constraint entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

// Get fetches the formation constraint from the db by the provided id
func (r *repository) Get(ctx context.Context, id string) (*model.FormationConstraint, error) {
	var entity Entity
	if err := r.singleGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	result := r.conv.FromEntity(&entity)

	return result, nil
}

// ListAll lists all formation constraints
func (r *repository) ListAll(ctx context.Context) ([]*model.FormationConstraint, error) {
	var entities EntityCollection

	if err := r.lister.ListGlobal(ctx, &entities); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListByIDs lists all formation constraints whose id is in formationConstraintIDs
func (r *repository) ListByIDs(ctx context.Context, formationConstraintIDs []string) ([]*model.FormationConstraint, error) {
	var entities EntityCollection

	if err := r.lister.ListGlobal(ctx, &entities, repo.NewInConditionForStringValues(idColumn, formationConstraintIDs)); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// Delete deletes formation constraint from the database by id
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.deleter.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// ListMatchingFormationConstraints lists formationConstraints whose ID can be found in formationConstraintIDs or have constraint scope "global" that match on the join point location and matching details
func (r *repository) ListMatchingFormationConstraints(ctx context.Context, formationConstraintIDs []string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error) {
	var entityCollection EntityCollection

	formationTypeRelevanceConditions := []repo.Condition{
		repo.NewEqualCondition("constraint_scope", model.GlobalFormationConstraintScope),
	}

	if len(formationConstraintIDs) > 0 {
		formationTypeRelevanceConditions = append(formationTypeRelevanceConditions, repo.NewInConditionForStringValues("id", formationConstraintIDs))
	}

	conditions := repo.And(
		append(
			repo.ConditionTreesFromConditions(
				[]repo.Condition{
					repo.NewEqualCondition("target_operation", location.OperationName),
					repo.NewEqualCondition("constraint_type", location.ConstraintType),
					repo.NewEqualCondition("resource_type", details.resourceType),
					repo.NewEqualCondition("resource_subtype", details.resourceSubtype),
				},
			),
			repo.Or(repo.ConditionTreesFromConditions(
				formationTypeRelevanceConditions,
			)...),
		)...,
	)

	if err := r.conditionTreeLister.ListConditionTreeGlobal(ctx, resource.FormationConstraint, &entityCollection, conditions); err != nil {
		return nil, errors.Wrap(err, "while listing constraints")
	}
	return r.multipleFromEntities(entityCollection)
}

func (r *repository) multipleFromEntities(entities EntityCollection) ([]*model.FormationConstraint, error) {
	items := make([]*model.FormationConstraint, 0, len(entities))
	for _, ent := range entities {
		m := r.conv.FromEntity(&ent)
		items = append(items, m)
	}
	return items, nil
}
