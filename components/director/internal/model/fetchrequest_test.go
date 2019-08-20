package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFetchRequestInput_ToFetchRequest(t *testing.T) {
	// given
	mode := model.FetchModeSingle
	filter := "foofilter"
	timestamp := time.Now()
	testCases := []struct {
		Name         string
		InputID string
		InputReferenceObjectType model.FetchRequestReferenceObjectType
		InputReferenceObjectID string
		InputFRInput *model.FetchRequestInput
		Expected     *model.FetchRequest
	}{
		{
			Name: "All properties given",
			InputFRInput: &model.FetchRequestInput{
				URL: "foourl",
				Auth: &model.AuthInput{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:   &mode,
				Filter: &filter,
			},
			Expected: &model.FetchRequest{
				URL: "foourl",
				Auth: &model.Auth{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:   mode,
				Filter: &filter,
				Status: &model.FetchRequestStatus{
					Condition: model.FetchRequestStatusConditionInitial,
					Timestamp: timestamp,
				},
			},
		},
		{
			Name:         "Empty",
			InputFRInput: &model.FetchRequestInput{},
			Expected: &model.FetchRequest{
				Mode: model.FetchModeSingle,
				Status: &model.FetchRequestStatus{
					Condition: model.FetchRequestStatusConditionInitial,
					Timestamp: timestamp,
				},
			},
		},
		{
			Name:         "Nil",
			InputFRInput: nil,
			Expected:     nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result := testCase.InputFRInput.ToFetchRequest(timestamp, testCase.InputID,  testCase.InputReferenceObjectType, testCase.InputReferenceObjectID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
