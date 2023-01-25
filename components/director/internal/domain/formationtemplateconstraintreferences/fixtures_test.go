package formationtemplateconstraintreferences_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	constraintID = "constraint_id"
	templateID   = "template_id"
)

func fixColumns() []string {
	return []string{"formation_template_id", "formation_constraint_id"}
}

var (
	gqlConstraintReference = &graphql.ConstraintReference{
		ConstraintID:        constraintID,
		FormationTemplateID: templateID,
	}
	constraintReference = &model.FormationTemplateConstraintReference{
		ConstraintID:        constraintID,
		FormationTemplateID: templateID,
	}
	nilModel *model.FormationTemplateConstraintReference
	entity   = &formationtemplateconstraintreferences.Entity{
		ConstraintID:        constraintID,
		FormationTemplateID: templateID,
	}
)

func unusedConstraintReferenceService() *automock.ConstraintReferenceService {
	return &automock.ConstraintReferenceService{}
}

func unusedConstraintReferenceConverter() *automock.ConstraintReferenceConverter {
	return &automock.ConstraintReferenceConverter{}
}
