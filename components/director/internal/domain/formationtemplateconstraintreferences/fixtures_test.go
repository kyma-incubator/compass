package formationtemplateconstraintreferences_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	constraintID = "constraint_id"
	templateID   = "template_id"
)

func fixColumns() []string {
	return []string{"formation_template", "formation_constraint"}
}

var (
	constraintReference = &model.FormationTemplateConstraintReference{
		Constraint:        constraintID,
		FormationTemplate: templateID,
	}
	nilModel *model.FormationTemplateConstraintReference
	entity   = &formationtemplateconstraintreferences.Entity{
		Constraint:        constraintID,
		FormationTemplate: templateID,
	}
)
