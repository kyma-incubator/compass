package assignmentoperation_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var converter = assignmentoperation.NewConverter()

func TestToGraphQL(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *model.AssignmentOperation
		Expected       *graphql.AssignmentOperation
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    fixAssignmentOperationModel(),
			Expected: fixAssignmentOperationGQL(),
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
			r := converter.ToGraphQL(testCase.Input)

			require.Equal(t, testCase.Expected, r)
		})
	}
}

func Test_converter_MultipleToGraphQL(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          []*model.AssignmentOperation
		Expected       []*graphql.AssignmentOperation
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    []*model.AssignmentOperation{fixAssignmentOperationModel()},
			Expected: []*graphql.AssignmentOperation{fixAssignmentOperationGQL()},
		},
		{
			Name:     "Success when input is empty slice",
			Input:    []*model.AssignmentOperation{},
			Expected: nil,
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
			r := converter.MultipleToGraphQL(testCase.Input)

			require.Equal(t, testCase.Expected, r)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *model.AssignmentOperation
		Expected       *assignmentoperation.Entity
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    fixAssignmentOperationModel(),
			Expected: fixAssignmentOperationEntity(),
		},
		{
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r := converter.ToEntity(testCase.Input)

			require.Equal(t, testCase.Expected, r)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *assignmentoperation.Entity
		Expected       *model.AssignmentOperation
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    fixAssignmentOperationEntity(),
			Expected: fixAssignmentOperationModel(),
		},
		{
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r := converter.FromEntity(testCase.Input)

			require.Equal(t, testCase.Expected, r)
		})
	}
}
