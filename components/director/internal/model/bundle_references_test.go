package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBundleReferenceInput_ToBundleReference(t *testing.T) {
	// given
	bundleID := "bundle-id"
	refID := "ref-id"
	tenant := "tnt"
	apiDefaultTargetURL := "http://test.com"
	testCases := []struct {
		Name                     string
		InputReferenceObjectType model.BundleReferenceObjectType
		InputReferenceObjectID   *string
		BundleReferenceInput     *model.BundleReferenceInput
		Expected                 *model.BundleReference
		ExpectedErr              error
	}{
		{
			Name:                     "All properties given for APIBundleReference",
			InputReferenceObjectID:   &refID,
			InputReferenceObjectType: model.BundleAPIReference,
			BundleReferenceInput: &model.BundleReferenceInput{
				APIDefaultTargetURL: &apiDefaultTargetURL,
			},
			Expected: &model.BundleReference{
				Tenant:              tenant,
				BundleID:            &bundleID,
				ObjectType:          model.BundleAPIReference,
				ObjectID:            &refID,
				APIDefaultTargetURL: &apiDefaultTargetURL,
			},
		},
		{
			Name:                     "Default targetURL for API is missing",
			InputReferenceObjectID:   &refID,
			InputReferenceObjectType: model.BundleAPIReference,
			BundleReferenceInput: &model.BundleReferenceInput{
				APIDefaultTargetURL: nil,
			},
			Expected:    nil,
			ExpectedErr: errors.New("default targetURL for API cannot be empty"),
		},
		{
			Name:                     "All properties given for EventBundleReference",
			InputReferenceObjectID:   &refID,
			InputReferenceObjectType: model.BundleEventReference,
			BundleReferenceInput: &model.BundleReferenceInput{
				APIDefaultTargetURL: nil,
			},
			Expected: &model.BundleReference{
				Tenant:              tenant,
				BundleID:            &bundleID,
				ObjectType:          model.BundleEventReference,
				ObjectID:            &refID,
				APIDefaultTargetURL: nil,
			},
		},
		{
			Name:                 "Nil",
			BundleReferenceInput: nil,
			Expected:             nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result, err := testCase.BundleReferenceInput.ToBundleReference(tenant, testCase.InputReferenceObjectType, &bundleID, testCase.InputReferenceObjectID)

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
