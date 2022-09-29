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
		Name     string
		Input    *model.FormationAssignment
		Expected *graphql.FormationAssignment
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
			Name:     "Error when configuration value is invalid json",
			Input:    fixFormationAssignmentModel(TestInvalidConfigValueRawJSON),
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r := converter.ToGraphQL(testCase.Input)

			require.Equal(t, r, testCase.Expected)
		})
	}
}

func Test_converter_MultipleToGraphQL(t *testing.T) {
	testCases := []struct {
		Name                         string
		InputFormationAssignments    []*model.FormationAssignment
		ExpectedFormationAssignments []*graphql.FormationAssignment
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := converter.MultipleToGraphQL(testCase.InputFormationAssignments)

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
