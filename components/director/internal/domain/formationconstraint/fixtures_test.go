package formationconstraint_test

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
	description              = "description"
	updatedDescription       = "updatedDescription"
	createdAt                = time.Now()
	priority                 = 2
	updatedPriority          = 7
	formationConstraintModel = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
		CreatedAt:       &createdAt,
	}
	formationConstraintModelUpdatedAllFields = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     updatedDescription,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        updatedPriority,
	}
	formationConstraintModelUpdatedTemplateAndPriority = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        updatedPriority,
	}
	formationConstraintModelUpdatedTemplateAndDescription = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     updatedDescription,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
	}
	gqlFormationConstraint = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
		Priority:        priority,
		CreatedAt:       graphql.Timestamp(createdAt),
	}
	gqlFormationConstraintUpdated = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
		Priority:        priority,
		CreatedAt:       graphql.Timestamp(createdAt),
	}
	gqlFormationConstraintUpdatedTemplateAndPriority = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
		Priority:        updatedPriority,
		CreatedAt:       graphql.Timestamp(createdAt),
	}
	gqlFormationConstraintUpdatedTemplateAndDescription = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     updatedDescription,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
		Priority:        priority,
		CreatedAt:       graphql.Timestamp(createdAt),
	}
	formationConstraintModel2 = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PostOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
		CreatedAt:       &createdAt,
	}
	gqlFormationConstraint2 = &graphql.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  string(model.PostOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
		Priority:        priority,
		CreatedAt:       graphql.Timestamp(createdAt),
	}
	formationConstraintInput = graphql.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     &description,
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: graphql.ConstraintScopeFormationType,
		Priority:        &priority,
	}
	formationConstraintInputUpdatedAllFields = graphql.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     &updatedDescription,
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: graphql.ConstraintScopeFormationType,
		Priority:        &updatedPriority,
	}
	formationConstraintInputUpdatedTemplateAndPriority = graphql.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     &description,
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: graphql.ConstraintScopeFormationType,
		Priority:        &updatedPriority,
	}
	formationConstraintInputUpdatedTemplateAndDescription = graphql.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     &updatedDescription,
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        operatorName,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplateUpdated,
		ConstraintScope: graphql.ConstraintScopeFormationType,
		Priority:        &priority,
	}
	formationConstraintModelInput = &model.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
	}
	formationConstraintModelInputUpdatedAllFields = &model.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     updatedDescription,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        updatedPriority,
	}
	formationConstraintModelInputUpdatedTemplateAndPriority = &model.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        updatedPriority,
	}
	formationConstraintModelInputUpdatedTemplateAndDescription = &model.FormationConstraintInput{
		Name:            formationConstraintName,
		Description:     updatedDescription,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
	}
	formationConstraintUpdateInput = graphql.FormationConstraintUpdateInput{
		InputTemplate: inputTemplateUpdated,
		Priority:      &updatedPriority,
		Description:   str.Ptr(updatedDescription),
	}
	formationConstraintUpdateInputWithTemplateAndPriority = graphql.FormationConstraintUpdateInput{
		InputTemplate: inputTemplateUpdated,
		Priority:      &updatedPriority,
	}
	formationConstraintUpdateInputWithTemplateAndDescription = graphql.FormationConstraintUpdateInput{
		InputTemplate: inputTemplateUpdated,
		Description:   str.Ptr(updatedDescription),
	}
	entity = formationconstraint.Entity{
		ID:              testID,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  string(model.PreOperation),
		TargetOperation: string(model.AssignFormationOperation),
		Operator:        operatorName,
		ResourceType:    string(model.ApplicationResourceType),
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: string(model.FormationTypeFormationConstraintScope),
		Priority:        priority,
		CreatedAt:       &createdAt,
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
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
	}
	formationConstraint2 = &model.FormationConstraint{
		ID:              constraintID2,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
		Priority:        priority,
	}
	globalConstraint = &model.FormationConstraint{
		ID:              constraintID3,
		Name:            formationConstraintName,
		Description:     description,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.GlobalFormationConstraintScope,
		Priority:        priority,
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
	return []string{"id", "name", "description", "constraint_type", "target_operation", "operator", "resource_type", "resource_subtype", "input_template", "constraint_scope", "priority", "created_at"}
}
