package formationconstraint

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.formation_constraints`

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
	lister repo.ConditionTreeListerGlobal
	conv   EntityConverter
}

// NewRepository creates a new FormationConstraint repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		lister: repo.NewConditionTreeListerGlobal(tableName, tableColumns),
		conv:   conv,
	}
}

// ListMatchingFormationConstraints lists formationConstraints whose ID can be found in formationConstraintIDs or have constraint scope "global" that match on the join point location and matching details
func (r *repository) ListMatchingFormationConstraints(ctx context.Context, formationConstraintIDs []string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error) {
	var entityCollection EntityCollection
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
				[]repo.Condition{
					repo.NewInConditionForStringValues("id", formationConstraintIDs),
					repo.NewEqualCondition("constraint_scope", model.GlobalFormationConstraintScope),
				},
			)...),
		)...,
	)

	if err := r.lister.ListConditionTreeGlobal(ctx, resource.FormationConstraint, &entityCollection, conditions); err != nil {
		return nil, errors.Wrap(err, "while listing constraints")
	}
	return r.conv.MultipleFromEntity(entityCollection), nil
}
