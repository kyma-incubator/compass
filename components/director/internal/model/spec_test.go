package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSpecInput_ToSpec(t *testing.T) {
	// given
	id := "id"
	refID := "ref-id"
	tenant := "tnt"
	data := "data"
	apiType := model.APISpecTypeOdata
	eventType := model.EventSpecTypeAsyncAPI
	testCases := []struct {
		Name                     string
		InputID                  string
		InputReferenceObjectType model.SpecReferenceObjectType
		InputReferenceObjectID   string
		SpecInput                *model.SpecInput
		Expected                 *model.Spec
		ExpectedErr              error
	}{
		{
			Name:                     "All properties given for API",
			InputID:                  id,
			InputReferenceObjectID:   refID,
			InputReferenceObjectType: model.APISpecReference,
			SpecInput: &model.SpecInput{
				Data:    &data,
				Format:  model.SpecFormatJSON,
				APIType: &apiType,
			},
			Expected: &model.Spec{
				ID:         id,
				Tenant:     tenant,
				ObjectType: model.APISpecReference,
				ObjectID:   refID,
				Data:       &data,
				Format:     model.SpecFormatJSON,
				APIType:    &apiType,
			},
		},
		{
			Name:                     "API Type missing",
			InputID:                  id,
			InputReferenceObjectID:   refID,
			InputReferenceObjectType: model.APISpecReference,
			SpecInput: &model.SpecInput{
				Data:   &data,
				Format: model.SpecFormatJSON,
			},
			Expected:    nil,
			ExpectedErr: errors.New("API Spec type cannot be empty"),
		},
		{
			Name:                     "All properties given for Event",
			InputID:                  id,
			InputReferenceObjectID:   refID,
			InputReferenceObjectType: model.EventSpecReference,
			SpecInput: &model.SpecInput{
				Data:      &data,
				Format:    model.SpecFormatJSON,
				EventType: &eventType,
			},
			Expected: &model.Spec{
				ID:         id,
				Tenant:     tenant,
				ObjectType: model.EventSpecReference,
				ObjectID:   refID,
				Data:       &data,
				Format:     model.SpecFormatJSON,
				EventType:  &eventType,
			},
		},
		{
			Name:                     "Event Type missing",
			InputID:                  id,
			InputReferenceObjectID:   refID,
			InputReferenceObjectType: model.EventSpecReference,
			SpecInput: &model.SpecInput{
				Data:   &data,
				Format: model.SpecFormatJSON,
			},
			Expected:    nil,
			ExpectedErr: errors.New("event spec type cannot be empty"),
		},
		{
			Name:      "Nil",
			SpecInput: nil,
			Expected:  nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result, err := testCase.SpecInput.ToSpec(testCase.InputID, tenant, testCase.InputReferenceObjectType, testCase.InputReferenceObjectID)

			// then
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.Expected, result)
		})
	}
}
