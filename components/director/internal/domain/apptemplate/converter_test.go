package apptemplate_test

import (
	"database/sql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToEntity(t *testing.T) {
	// given
	id := "foo"
	name := "bar"

	appTemplateEntity := fixEntityAppTemplate(t, id, name)
	appTemplateModel := fixModelAppTemplate(id, name)

	testCases := []struct {
		Name     string
		Input    *model.ApplicationTemplate
		Expected *apptemplate.Entity
	}{
		{
			Name:     "All properties given",
			Input:    appTemplateModel,
			Expected: appTemplateEntity,
		},
		{
			Name:     "Empty",
			Input:    &model.ApplicationTemplate{},
			Expected: &apptemplate.Entity{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
		// Cannot test error case while marshalling Application Input or Placeholders: cannot create invalid object
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apptemplate.NewConverter(nil)

			// when
			res, err := conv.ToEntity(testCase.Input)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// given
	id := "foo"
	name := "bar"

	appTemplateEntity := fixEntityAppTemplate(t, id, name)
	appTemplateModel := fixModelAppTemplate(id, name)

	testCases := []struct {
		Name               string
		Input              *apptemplate.Entity
		Expected           *model.ApplicationTemplate
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
			Input:              &apptemplate.Entity{},
			Expected:           &model.ApplicationTemplate{},
			ExpectedErrMessage: "",
		},
		{
			Name:               "Nil",
			Input:              nil,
			Expected:           nil,
			ExpectedErrMessage: "",
		},
		{
			Name: "Application Input Unmarshall Error",
			Input: &apptemplate.Entity{
				ApplicationInput: "{dasdd",
			},
			ExpectedErrMessage: "while unpacking Application Create Input: invalid character 'd' looking for beginning of object key string",
		},
		{
			Name: "Placeholders Unmarshall Error",
			Input: &apptemplate.Entity{
				Placeholders: sql.NullString{
					String: "{dasdd",
					Valid:  true,
				},
			},
			ExpectedErrMessage: "while unpacking Placeholders: invalid character 'd' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apptemplate.NewConverter(nil)

			// when
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
