package formationtemplateconstraintreferences

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
	tableName                string = `public.formation_template_constraint_references`
	formationTemplateField   string = "formation_template"
	formationConstraintField string = "constraint"
)

var (
	tableColumns = []string{formationConstraintField, formationTemplateField}
)

// EntityConverter converts between the internal model and entity
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.FormationTemplateConstraintReference) *Entity
	FromEntity(entity *Entity) *model.FormationTemplateConstraintReference
	MultipleFromEntity(in EntityCollection) []*model.FormationTemplateConstraintReference
}

type repository struct {
	lister  repo.ListerGlobal
	creator repo.CreatorGlobal
	deleter repo.DeleterGlobal
	conv    EntityConverter
}

// NewRepository creates a new FormationConstraint repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		lister:  repo.NewListerGlobal(resource.FormationTemplateConstraintReference, tableName, tableColumns),
		creator: repo.NewCreatorGlobal(resource.FormationTemplateConstraintReference, tableName, tableColumns),
		deleter: repo.NewDeleterGlobal(resource.FormationTemplateConstraintReference, tableName),
		conv:    conv,
	}
}

// ListByFormationTemplateID lists formationConstraints which name can be found in constraintNames and formation constraint which have constraint scope "global"
func (r *repository) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationTemplateConstraintReference, error) {
	var entityCollection EntityCollection

	if err := r.lister.ListGlobal(ctx, &entityCollection, repo.NewEqualCondition(formationTemplateField, formationTemplateID)); err != nil {
		//TODO do we need to check for not found?
		return nil, errors.Wrap(err, "while listing formationTemplate-constraint references by formationTemplate ID")
	}
	return r.conv.MultipleFromEntity(entityCollection), nil
}

// Create stores new formationTemplate-constraint reference in the database
func (r *repository) Create(ctx context.Context, item *model.FormationTemplateConstraintReference) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting FormationTemplateConstraintReference with formationTemplate ID: %q and formationConstraint ID: %q to entity", item.FormationTemplate, item.Constraint)
	entity := r.conv.ToEntity(item)

	log.C(ctx).Debugf("Persisting FormationTemplateConstraintReference with formationTemplate ID: %q and formationConstraint ID: %q to the DB", item.FormationTemplate, item.Constraint)
	return r.creator.Create(ctx, entity)
}

// Delete deletes a formationTemplate-constraint reference for formation and constraint
func (r *repository) Delete(ctx context.Context, formationTemplateID, constraintID string) error {
	log.C(ctx).Debugf("Deleting FormationTemplateConstraintReference with formationTemplate ID: %q and formationConstraint ID: %q...", formationTemplateID, constraintID)
	return r.deleter.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(formationTemplateField, formationTemplateID), repo.NewEqualCondition(formationConstraintField, constraintID)})
}
