package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFetchRequestInput_ToFetchRequest(t *testing.T) {
	// GIVEN
	mode := model.FetchModeSingle
	filter := "foofilter"
	timestamp := time.Now()
	testCases := []struct {
		Name                     string
		InputID                  string
		InputReferenceObjectType model.FetchRequestReferenceObjectType
		InputReferenceObjectID   string
		InputFRInput             *model.FetchRequestInput
		Expected                 *model.FetchRequest
	}{
		{
			Name:                     "All properties given",
			InputID:                  "input-id",
			InputReferenceObjectID:   "ref-id",
			InputReferenceObjectType: model.APISpecFetchRequestReference,
			InputFRInput: &model.FetchRequestInput{
				URL: "foourl",
				Auth: &auth.AuthInput{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:   &mode,
				Filter: &filter,
			},
			Expected: &model.FetchRequest{
				ID:         "input-id",
				ObjectID:   "ref-id",
				ObjectType: model.APISpecFetchRequestReference,
				URL:        "foourl",
				Auth: &auth.Auth{
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
			Name:                     "Empty",
			InputID:                  "input-id",
			InputReferenceObjectType: model.APISpecFetchRequestReference,
			InputReferenceObjectID:   "ref-id-2",
			InputFRInput:             &model.FetchRequestInput{},
			Expected: &model.FetchRequest{
				ID:         "input-id",
				ObjectID:   "ref-id-2",
				ObjectType: model.APISpecFetchRequestReference,
				Mode:       model.FetchModeSingle,
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
			// WHEN
			result := testCase.InputFRInput.ToFetchRequest(timestamp, testCase.InputID, testCase.InputReferenceObjectType, testCase.InputReferenceObjectID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
