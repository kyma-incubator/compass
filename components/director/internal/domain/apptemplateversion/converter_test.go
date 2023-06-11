package apptemplateversion_test

import (
	"database/sql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockedError = errors.New("test-error")

func TestConverter_ToEntity(t *testing.T) {
	// GIVEN
	appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
	appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

	testCases := []struct {
		Name     string
		Input    *model.ApplicationTemplateVersion
		Expected *apptemplateversion.Entity
	}{
		{
			Name:     "All properties given",
			Input:    appTemplateVersionModel,
			Expected: appTemplateVersionEntity,
		},
		{
			Name:     "Empty",
			Input:    &model.ApplicationTemplateVersion{},
			Expected: &apptemplateversion.Entity{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apptemplateversion.NewConverter()

			// WHEN
			res, err := conv.ToEntity(testCase.Input)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// GIVEN

	appTemplateEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)
	appTemplateModel := fixModelApplicationTemplateVersion(appTemplateVersionID)

	testCases := []struct {
		Name               string
		Input              *apptemplateversion.Entity
		Expected           *model.ApplicationTemplateVersion
		ExpectedErrMessage string
	}{
		{
			Name:               "All properties given",
			Input:              appTemplateEntity,
			Expected:           appTemplateModel,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Empty",
			Input:              &apptemplateversion.Entity{},
			Expected:           &model.ApplicationTemplateVersion{},
			ExpectedErrMessage: "",
		},
		{
			Name:               "Nil",
			Input:              nil,
			Expected:           nil,
			ExpectedErrMessage: "",
		},
		{
			Name: "PlaceholdersJSON Unmarshall Error",
			Input: &apptemplateversion.Entity{
				CorrelationIDs: sql.NullString{
					String: "{dasdd",
					Valid:  true,
				},
			},
			ExpectedErrMessage: "while converting correlationIDs to string array: invalid character 'd' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apptemplateversion.NewConverter()

			// WHEN
			res, err := conv.FromEntity(testCase.Input)

			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
			} else {
				require.Nil(t, err)
			}

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
