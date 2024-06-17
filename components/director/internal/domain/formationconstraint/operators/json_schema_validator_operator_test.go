package operators_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_JSONSchemaValidator(t *testing.T) {
	JSONSchemaOperatorInputWithExceptFormationType.WithExceptFormationTypes([]string{formationTemplateName})
	JSONSchemaOperatorInputWithExceptSubtypes.WithExceptSubtypes([]string{resourceSubtype})
	JSONSchemaOperatorInputWithNonExistingOnlyForSourceSubtypes.WithOnlyForSourceSubtypes([]string{"non-existing-source-subtype"})
	JSONSchemaOperatorInputWithOnlyForSourceSubtypes.WithOnlyForSourceSubtypes([]string{resourceSubtype})

	testCases := []struct {
		Name             string
		Input            operators.OperatorInput
		LabelSvc         func() *automock.LabelService
		ExpectedResult   bool
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success with provided JSON schema",
			Input:          JSONSchemaOperatorInput,
			ExpectedResult: true,
		},
		{
			Name:             "Error when parsing operator input",
			Input:            "wrong input",
			ExpectedErrorMsg: fmt.Sprintf("Incompatible input for operator: %s", operators.JSONSchemaValidatorOperator),
		},
		{
			Name:             "Error when retrieving formation assignment pointer fails",
			Input:            JSONSchemaOperatorInputWithoutAssignmentMemoryAddress,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:           "Success(no-op) when there is no configuration in the formation assignment",
			Input:          JSONSchemaOperatorInputWithEmptyAssignmentConfig,
			ExpectedResult: true,
		},
		{
			Name:             "Error when the assignment has configuration but the JSON schema is empty",
			Input:            JSONSchemaOperatorInputWithEmptySchema,
			ExpectedErrorMsg: "could not be validated due to empty JSON schema.",
		},
		{
			Name:             "Error when the provided JSON schema is invalid",
			Input:            JSONSchemaOperatorInputWithInvalidSchema,
			ExpectedErrorMsg: "while creating JSON Schema validator",
		},
		{
			Name:             "Error when the provided assignment config is not valid against the JSON schema",
			Input:            JSONSchemaOperatorInputWithAssignmentConfigInvalidAgainstSchema,
			ExpectedErrorMsg: "while validating the initial config against the JSON Schema",
		},
		{
			Name:           "Success(no-op) when the formation type is excluded from validation",
			Input:          JSONSchemaOperatorInputWithExceptFormationType,
			ExpectedResult: true,
		},
		{
			Name:  "Success(no-op) when the source resource type is not part of the only source subtypes configuration",
			Input: JSONSchemaOperatorInputWithExceptSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(&model.Label{Value: resourceSubtype}, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Error when getting subtype of source resource fails",
			Input: JSONSchemaOperatorInputWithExceptSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrorMsg: fmt.Sprintf("while getting subtype of resource with type: %s and ID: %s", model.ApplicationResourceType, sourceAppID),
		},
		{
			Name:  "Success(no-op) when the source resource type is NOT part of the only source subtypes configuration",
			Input: JSONSchemaOperatorInputWithNonExistingOnlyForSourceSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(&model.Label{Value: resourceSubtype}, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success when the source resource type is part of the only source subtypes configuration",
			Input: JSONSchemaOperatorInputWithOnlyForSourceSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(&model.Label{Value: resourceSubtype}, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := unusedLabelService()
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}
			defer mock.AssertExpectationsForObjects(t, labelSvc)

			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, labelSvc, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			result, err := engine.JSONSchemaValidator(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, labelSvc)
		})
	}
}
