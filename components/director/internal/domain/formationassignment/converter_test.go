package formationassignment_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

var converter = formationassignment.NewConverter()

func TestToGraphQL(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *model.FormationAssignment
		Expected       *graphql.FormationAssignment
		ExpectedErrMsg string
	}{
		{
			Name:     "Success when assignment contains configuration",
			Input:    fixFormationAssignmentModel(TestConfigValueRawJSON),
			Expected: fixFormationAssignmentGQLModel(&TestConfigValueStr),
		},
		{
			Name:     "Success when assignment contains error",
			Input:    fixFormationAssignmentModelWithError(TestErrorValueRawJSON),
			Expected: fixFormationAssignmentGQLModelWithError(&TestErrorValueStr),
		},
		{
			Name:     "Success when assignment contains configuration and error",
			Input:    fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON),
			Expected: fixFormationAssignmentGQLModelWithConfigAndError(&TestConfigValueStr, &TestErrorValueStr),
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
		{
			Name:           "Error when configuration value is invalid json",
			Input:          fixFormationAssignmentModel(TestInvalidConfigValueRawJSON),
			Expected:       nil,
			ExpectedErrMsg: "while converting formation assignment to GraphQL",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r, err := converter.ToGraphQL(testCase.Input)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, r, testCase.Expected)
		})
	}
}

func Test_converter_MultipleToGraphQL(t *testing.T) {
	testCases := []struct {
		Name                         string
		InputFormationAssignments    []*model.FormationAssignment
		ExpectedFormationAssignments []*graphql.FormationAssignment
		ExpectedErrorMsg             string
	}{
		{
			Name:                         "Success when assignment contains configuration",
			InputFormationAssignments:    []*model.FormationAssignment{fixFormationAssignmentModel(TestConfigValueRawJSON)},
			ExpectedFormationAssignments: []*graphql.FormationAssignment{fixFormationAssignmentGQLModel(&TestConfigValueStr)},
		},
		{
			Name:                         "Success when assignment contains error",
			InputFormationAssignments:    []*model.FormationAssignment{fixFormationAssignmentModelWithError(TestErrorValueRawJSON)},
			ExpectedFormationAssignments: []*graphql.FormationAssignment{fixFormationAssignmentGQLModelWithError(&TestErrorValueStr)},
		},
		{
			Name:                         "Success when assignment contains configuration and error",
			InputFormationAssignments:    []*model.FormationAssignment{fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON)},
			ExpectedFormationAssignments: []*graphql.FormationAssignment{fixFormationAssignmentGQLModelWithConfigAndError(&TestConfigValueStr, &TestErrorValueStr)},
		},
		{
			Name:                         "Success when input is nil",
			InputFormationAssignments:    nil,
			ExpectedFormationAssignments: nil,
		},
		{
			Name:                         "Success when one of the input formations is nil",
			InputFormationAssignments:    []*model.FormationAssignment{nil, fixFormationAssignmentModel(TestConfigValueRawJSON)},
			ExpectedFormationAssignments: []*graphql.FormationAssignment{fixFormationAssignmentGQLModel(&TestConfigValueStr)},
		},
		{
			Name:                         "Error when conversion to GraphQL failed",
			InputFormationAssignments:    []*model.FormationAssignment{fixFormationAssignmentModel(TestInvalidConfigValueRawJSON)},
			ExpectedFormationAssignments: []*graphql.FormationAssignment{},
			ExpectedErrorMsg:             "while converting formation assignment to GraphQL",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result, err := converter.MultipleToGraphQL(testCase.InputFormationAssignments)
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, testCase.ExpectedFormationAssignments, result)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.FormationAssignment
		Expected *formationassignment.Entity
	}{
		{
			Name:     "Success when assignment contains configuration",
			Input:    fixFormationAssignmentModel(TestConfigValueRawJSON),
			Expected: fixFormationAssignmentEntity(TestConfigValueStr),
		},
		{
			Name:     "Success when assignment contains error",
			Input:    fixFormationAssignmentModelWithError(TestErrorValueRawJSON),
			Expected: fixFormationAssignmentEntityWithError(TestErrorValueStr),
		},
		{
			Name:     "Success when assignment contains configuration and error",
			Input:    fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON),
			Expected: fixFormationAssignmentEntityWithConfigurationAndError(TestConfigValueStr, TestErrorValueStr),
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := converter.ToEntity(testCase.Input)

			require.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *formationassignment.Entity
		Expected *model.FormationAssignment
	}{
		{
			Name:     "Success when assignment contains configuration",
			Input:    fixFormationAssignmentEntity(TestConfigValueStr),
			Expected: fixFormationAssignmentModel(TestConfigValueRawJSON),
		},
		{
			Name:     "Success when assignment contains error",
			Input:    fixFormationAssignmentEntityWithError(TestErrorValueStr),
			Expected: fixFormationAssignmentModelWithError(TestErrorValueRawJSON),
		},
		{
			Name:     "Success when assignment contains configuration and error",
			Input:    fixFormationAssignmentEntityWithConfigurationAndError(TestConfigValueStr, TestErrorValueStr),
			Expected: fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON),
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := converter.FromEntity(testCase.Input)

			require.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_ToInput(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.FormationAssignment
		Expected *model.FormationAssignmentInput
	}{
		{
			Name:     "Success when assignment contains configuration",
			Input:    fixFormationAssignmentModel(TestConfigValueRawJSON),
			Expected: fixFormationAssignmentModelInput(TestConfigValueRawJSON),
		},
		{
			Name:     "Success when assignment contains error",
			Input:    fixFormationAssignmentModelWithError(TestErrorValueRawJSON),
			Expected: fixFormationAssignmentModelInputWithError(TestErrorValueRawJSON),
		},
		{
			Name:     "Success when assignment contains configuration and error",
			Input:    fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON),
			Expected: fixFormationAssignmentModelInputWithConfigurationAndError(TestConfigValueRawJSON, TestErrorValueRawJSON),
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := converter.ToInput(testCase.Input)

			require.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_FromInput(t *testing.T) {
	formationAssignmentWithConfig := fixFormationAssignmentModel(TestConfigValueRawJSON)
	formationAssignmentWithConfig.TenantID = ""
	formationAssignmentWithConfig.ID = ""

	formationAssignmentWithError := fixFormationAssignmentModelWithError(TestErrorValueRawJSON)
	formationAssignmentWithError.TenantID = ""
	formationAssignmentWithError.ID = ""

	formationAssignmentWithConfigAndError := fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON)
	formationAssignmentWithConfigAndError.TenantID = ""
	formationAssignmentWithConfigAndError.ID = ""

	testCases := []struct {
		Name     string
		Input    *model.FormationAssignmentInput
		Expected *model.FormationAssignment
	}{
		{
			Name:     "Success when assignment contains configuration",
			Input:    fixFormationAssignmentModelInput(TestConfigValueRawJSON),
			Expected: formationAssignmentWithConfig,
		},
		{
			Name:     "Success when assignment contains error",
			Input:    fixFormationAssignmentModelInputWithError(TestErrorValueRawJSON),
			Expected: formationAssignmentWithError,
		},
		{
			Name:     "Success when assignment contains configuration and error",
			Input:    fixFormationAssignmentModelInputWithConfigurationAndError(TestConfigValueRawJSON, TestErrorValueRawJSON),
			Expected: formationAssignmentWithConfigAndError,
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := converter.FromInput(testCase.Input)

			require.Equal(t, result, testCase.Expected)
		})
	}
}
