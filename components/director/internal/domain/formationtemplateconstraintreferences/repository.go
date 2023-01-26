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
	tableName                 string = `public.formation_template_constraint_references`
	formationTemplateColumn   string = "formation_template_id"
	formationConstraintColumn string = "formation_constraint_id"
)

var (
	tableColumns = []string{formationConstraintColumn, formationTemplateColumn}
)

type repository struct {
	lister  repo.ListerGlobal
	creator repo.CreatorGlobal
	deleter repo.DeleterGlobal
	conv    constraintReferenceConverter
}

// NewRepository creates a new FormationTemplateConstraintReference repository
func NewRepository(conv constraintReferenceConverter) *repository {
	return &repository{
		lister:  repo.NewListerGlobal(resource.FormationTemplateConstraintReference, tableName, tableColumns),
		creator: repo.NewCreatorGlobal(resource.FormationTemplateConstraintReference, tableName, tableColumns),
		deleter: repo.NewDeleterGlobal(resource.FormationTemplateConstraintReference, tableName),
		conv:    conv,
	}
}

// ListByFormationTemplateID lists formationTemplateConstraintReferences for the provided formationTemplate ID
func (r *repository) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationTemplateConstraintReference, error) {
	var entityCollection EntityCollection

	if err := r.lister.ListGlobal(ctx, &entityCollection, repo.NewEqualCondition(formationTemplateColumn, formationTemplateID)); err != nil {
		return nil, errors.Wrap(err, "while listing formationTemplate-constraint references by formationTemplate ID")
	}
	return r.multipleFromEntities(entityCollection)
}

// Create stores new formationTemplateConstraintReference in the database
func (r *repository) Create(ctx context.Context, item *model.FormationTemplateConstraintReference) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting FormationTemplateConstraintReference with formationTemplate ID: %q and formationConstraint ID: %q to entity", item.FormationTemplateID, item.ConstraintID)
	entity := r.conv.ToEntity(item)

	log.C(ctx).Debugf("Persisting FormationTemplateConstraintReference with formationTemplate ID: %q and formationConstraint ID: %q to the DB", item.FormationTemplateID, item.ConstraintID)
	return r.creator.Create(ctx, entity)
}

// Delete deletes a formationTemplateConstraintReference for formationTemplate ID and constraint ID
func (r *repository) Delete(ctx context.Context, formationTemplateID, constraintID string) error {
	log.C(ctx).Debugf("Deleting FormationTemplateConstraintReference with formationTemplate ID: %q and formationConstraint ID: %q...", formationTemplateID, constraintID)
	return r.deleter.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(formationTemplateColumn, formationTemplateID), repo.NewEqualCondition(formationConstraintColumn, constraintID)})
}

func (r *repository) multipleFromEntities(entities EntityCollection) ([]*model.FormationTemplateConstraintReference, error) {
	items := make([]*model.FormationTemplateConstraintReference, 0, len(entities))
	for _, ent := range entities {
		m := r.conv.FromEntity(ent)
		items = append(items, m)
	}
	return items, nil
}
