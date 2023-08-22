package formationconstraint_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	testID                  = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	formationTemplateID     = "id"
	formationConstraintName = "test constraint"
	operatorName            = operators.IsNotAssignedToAnyFormationOfTypeOperator
	resourceSubtype         = "test subtype"
	resourceSubtypeANY      = "ANY"
	inputTemplate           = `{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}"}`
	inputTemplateUpdated    = `{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}", "newField": "value"}`
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
	formationConstraintModelUpdated = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
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
	gqlFormationConstraintUpdated = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
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
	formationConstraintInputUpdated = graphql.FormationConstraintInput{
		Name:            formationConstraintName,
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
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
	formationConstraintUpdateInput = graphql.FormationConstraintUpdateInput{
		InputTemplate: inputTemplateUpdated,
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
	location = formationconstraintpkg.JoinPointLocation{
		OperationName:  "assign",
		ConstraintType: "pre",
	}
	details = formationconstraintpkg.AssignFormationOperationDetails{
		ResourceType:    "runtime",
		ResourceSubtype: "kyma",
	}
	matchingDetails = details.GetMatchingDetails()

	gqlInput       = &graphql.FormationConstraintInput{Name: testName}
	modelInput     = &model.FormationConstraintInput{Name: testName}
	modelFromInput = &model.FormationConstraint{ID: testID, Name: testName}

	formationTemplateID1 = "123"
	formationTemplateID2 = "456"
	formationTemplateID3 = "789"
	constraintID1        = "constraintID1"
	constraintID2        = "constraintID2"
	constraintID3        = "constraintID3"

	formationConstraint1 = &model.FormationConstraint{
		ID:              constraintID1,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}
	formationConstraint2 = &model.FormationConstraint{
		ID:              constraintID2,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}
	globalConstraint = &model.FormationConstraint{
		ID:              constraintID3,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.GlobalFormationConstraintScope,
	}
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

func fixColumns() []string {
	return []string{"id", "name", "constraint_type", "target_operation", "operator", "resource_type", "resource_subtype", "input_template", "constraint_scope"}
}
