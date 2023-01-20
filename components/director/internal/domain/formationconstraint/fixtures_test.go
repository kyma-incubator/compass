package formationconstraint_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraint2 "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	testID                  = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	formationTemplateID     = "id"
	formationConstraintName = "test constraint"
	operatorName            = formationconstraint.IsNotAssignedToAnyFormationOfTypeOperator
	resourceSubtype         = "test subtype"
	inputTemplate           = `{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}"}`
	testTenantID            = "d9fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	testName                = "test"
)

var (
	formationConstraintModel = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}
	gqlFormationConstraint = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
	}
	formationConstraintModel2 = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  model.PostOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}
	gqlFormationConstraint2 = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  string(model.PostOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
	}
	formationConstraintInput = graphql.FormationConstraintInput{
		Name:            formationConstraintName,
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}
	formationConstraintModelInput = &model.FormationConstraintInput{
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}
	entity = formationconstraint.Entity{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
	}
	nilModelEntity               *model.FormationConstraint
	formationConstraintReference = &model.FormationTemplateConstraintReference{
		ConstraintID:        testID,
		FormationTemplateID: formationTemplateID,
	}
	location = formationconstraint2.JoinPointLocation{
		OperationName:  "assign",
		ConstraintType: "pre",
	}
	details = formationconstraint2.AssignFormationOperationDetails{
		ResourceType:    "runtime",
		ResourceSubtype: "kyma",
	}
	matchingDetails = details.GetMatchingDetails()

	gqlInput       = &graphql.FormationConstraintInput{Name: testName}
	modelInput     = &model.FormationConstraintInput{Name: testName}
	modelFromInput = &model.FormationConstraint{ID: testID, Name: testName}
)

func UnusedFormationConstraintService() *automock.FormationConstraintService {
	return &automock.FormationConstraintService{}
}

func UnusedFormationConstraintRepository() *automock.FormationConstraintRepository {
	return &automock.FormationConstraintRepository{}
}

func UnusedFormationConstraintConverter() *automock.FormationConstraintConverter {
	return &automock.FormationConstraintConverter{}
}

func UnusedFormationTemplateConstraintReferenceRepository() *automock.FormationTemplateConstraintReferenceRepository {
	return &automock.FormationTemplateConstraintReferenceRepository{}
}

func fixColumns() []string {
	return []string{"id", "name", "constraint_type", "target_operation", "operator", "resource_type", "resource_subtype", "operator_scope", "input_template", "constraint_scope"}
}
