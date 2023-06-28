package apptemplateversion_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

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
			res := conv.ToEntity(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// GIVEN

	appTemplateEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)
	appTemplateModel := fixModelApplicationTemplateVersion(appTemplateVersionID)

	testCases := []struct {
		Name     string
		Input    *apptemplateversion.Entity
		Expected *model.ApplicationTemplateVersion
	}{
		{
			Name:     "All properties given",
			Input:    appTemplateEntity,
			Expected: appTemplateModel,
		},
		{
			Name:     "Empty",
			Input:    &apptemplateversion.Entity{},
			Expected: &model.ApplicationTemplateVersion{},
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
			res := conv.FromEntity(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
