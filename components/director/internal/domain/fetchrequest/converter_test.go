package fetchrequest_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.FetchRequest
		Expected *graphql.FetchRequest
	}{
		{
			Name:     "All properties given",
			Input:    fixModelFetchRequest(t, "foo", "bar"),
			Expected: fixGQLFetchRequest(t, "foo", "bar"),
		},
		{
			Name:  "Empty",
			Input: &model.FetchRequest{},
			Expected: &graphql.FetchRequest{
				Status: &graphql.FetchRequestStatus{
					Condition: graphql.FetchRequestStatusConditionInitial,
				},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("ToGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := fetchrequest.NewConverter(authConv)

			// when
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.FetchRequestInput
		Expected *model.FetchRequestInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLFetchRequestInput("foo", "bar"),
			Expected: fixModelFetchRequestInput("foo", "bar"),
		},
		{
			Name:     "Empty",
			Input:    &graphql.FetchRequestInput{},
			Expected: &model.FetchRequestInput{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("InputFromGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := fetchrequest.NewConverter(authConv)

			// when
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}
