package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAPIDefinitionInput_ToAPIDefinition(t *testing.T) {
	// given
	id := "foo"
	appID := str.Ptr("bar")
	desc := "Sample"
	name := "sample"
	targetUrl := "https://foo.bar"
	group := "sampleGroup"
	tenant := "tenant"

	testCases := []struct {
		Name     string
		Input    *model.APIDefinitionInput
		Expected *model.APIDefinition
	}{
		{
			Name: "All properties given",
			Input: &model.APIDefinitionInput{
				Name:        name,
				Description: &desc,
				TargetURL:   targetUrl,
				Group:       &group,
			},
			Expected: &model.APIDefinition{
				ID:            id,
				ApplicationID: appID,
				PackageID:     nil,
				Name:          name,
				Description:   &desc,
				TargetURL:     targetUrl,
				Group:         &group,
				Tenant:        tenant,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToAPIDefinition(id, appID, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestAPIDefinitionInput_ToAPIDefinitionWithPackageID(t *testing.T) {
	// given
	id := "foo"
	pkgID := str.Ptr("bar")
	desc := "Sample"
	name := "sample"
	targetUrl := "https://foo.bar"
	group := "sampleGroup"
	tenant := "tenant"

	testCases := []struct {
		Name     string
		Input    *model.APIDefinitionInput
		Expected *model.APIDefinition
	}{
		{
			Name: "All properties given",
			Input: &model.APIDefinitionInput{
				Name:        name,
				Description: &desc,
				TargetURL:   targetUrl,
				Group:       &group,
			},
			Expected: &model.APIDefinition{
				ID:            id,
				ApplicationID: nil,
				PackageID:     pkgID,
				Name:          name,
				Description:   &desc,
				TargetURL:     targetUrl,
				Group:         &group,
				Tenant:        tenant,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToAPIDefinitionWithinPackage(id, pkgID, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestAPIDefinitionInput_ToAPISpec(t *testing.T) {
	// given
	data := "bar"
	format := model.SpecFormat("Sample")
	specType := model.APISpecType("sample")

	testCases := []struct {
		Name     string
		Input    *model.APISpecInput
		Expected *model.APISpec
	}{
		{
			Name: "All properties given",
			Input: &model.APISpecInput{
				Data:   &data,
				Format: format,
				Type:   specType,
			},
			Expected: &model.APISpec{
				Data:   &data,
				Format: format,
				Type:   specType,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToAPISpec()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
