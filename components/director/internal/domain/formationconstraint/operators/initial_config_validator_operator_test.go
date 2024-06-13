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

func TestConstraintOperators_InitialConfigValidator(t *testing.T) {
	initialConfigOperatorInputWithExceptFormationType.WithExceptFormationTypes([]string{formationTemplateName})
	initialConfigOperatorInputWithExceptSubtypes.WithExceptSubtypes([]string{resourceSubtype})
	initialConfigOperatorInputWithNonExistingOnlyForSourceSubtypes.WithOnlyForSourceSubtypes([]string{"non-existing-source-subtype"})
	initialConfigOperatorInputWithOnlyForSourceSubtypes.WithOnlyForSourceSubtypes([]string{resourceSubtype})

	testCases := []struct {
		Name             string
		Input            operators.OperatorInput
		LabelSvc         func() *automock.LabelService
		ExpectedResult   bool
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success with provided JSON schema",
			Input:          initialConfigOperatorInput,
			ExpectedResult: true,
		},
		{
			Name:             "Error when parsing operator input",
			Input:            "wrong input",
			ExpectedErrorMsg: fmt.Sprintf("Incompatible input for operator: %s", operators.InitialConfigValidatorOperator),
		},
		{
			Name:             "Error when retrieving formation assignment pointer fails",
			Input:            initialConfigOperatorInputWithoutAssignmentMemoryAddress,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:           "Success(no-op) when there is no configuration in the formation assignment",
			Input:          initialConfigOperatorInputWithEmptyAssignmentConfig,
			ExpectedResult: true,
		},
		{
			Name:             "Error when the assignment has configuration but the JSON schema is empty",
			Input:            initialConfigOperatorInputWithEmptyJSONSchema,
			ExpectedErrorMsg: "could not be validated due to empty JSON schema.",
		},
		{
			Name:             "Error when the provided JSON schema is invalid",
			Input:            initialConfigOperatorInputWithInvalidJSONSchema,
			ExpectedErrorMsg: "while creating JSON Schema validator",
		},
		{
			Name:             "Error when the provided assignment config is not valid against the JSON schema",
			Input:            initialConfigOperatorInputWithAssignmentConfigInvalidAgainstJSONSchema,
			ExpectedErrorMsg: "while validating the initial config against the JSON Schema",
		},
		{
			Name:           "Success(no-op) when the formation type is excluded from validation",
			Input:          initialConfigOperatorInputWithExceptFormationType,
			ExpectedResult: true,
		},
		{
			Name:  "Success(no-op) when the source resource type is not part of the only source subtypes configuration",
			Input: initialConfigOperatorInputWithExceptSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(&model.Label{Value: resourceSubtype}, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Error when getting subtype of source resource fails",
			Input: initialConfigOperatorInputWithExceptSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrorMsg: fmt.Sprintf("while getting subtype of resource with type: %s and ID: %s", model.ApplicationResourceType, sourceAppID),
		},
		{
			Name:  "Success(no-op) when the source resource type is NOT part of the only source subtypes configuration",
			Input: initialConfigOperatorInputWithNonExistingOnlyForSourceSubtypes,
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceAppID, applicationTypeLabel).Return(&model.Label{Value: resourceSubtype}, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success when the source resource type is part of the only source subtypes configuration",
			Input: initialConfigOperatorInputWithOnlyForSourceSubtypes,
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

			result, err := engine.ConfigSchemaValidator(ctx, testCase.Input)

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
