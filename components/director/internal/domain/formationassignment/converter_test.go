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
			Name:     "Success",
			Input:    fixFormationAssignmentModel(TestConfigValueRawJSON),
			Expected: fixFormationAssignmentGQLModel(&TestConfigValueStr),
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
			Name:                         "Success",
			InputFormationAssignments:    []*model.FormationAssignment{fixFormationAssignmentModel(TestConfigValueRawJSON)},
			ExpectedFormationAssignments: []*graphql.FormationAssignment{fixFormationAssignmentGQLModel(&TestConfigValueStr)},
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
			Name:     "Success",
			Input:    fixFormationAssignmentModel(TestConfigValueRawJSON),
			Expected: fixFormationAssignmentEntity(TestConfigValueStr),
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
			Name:     "Success",
			Input:    fixFormationAssignmentEntity(TestConfigValueStr),
			Expected: fixFormationAssignmentModel(TestConfigValueRawJSON),
		}, {
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
