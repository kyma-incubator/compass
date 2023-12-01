package integrationdependency_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
		require.NotNil(t, integrationDependencyModel)
		conv := integrationdependency.NewConverter(version.NewConverter(), aspect.NewConverter())

		entity := conv.ToEntity(integrationDependencyModel)

		assert.Equal(t, fixIntegrationDependencyEntity(integrationDependencyID, appID), entity)
	})

	t.Run("Returns nil if integration dependency model is nil", func(t *testing.T) {
		conv := integrationdependency.NewConverter(version.NewConverter(), aspect.NewConverter())

		integrationDependencyEntity := conv.ToEntity(nil)

		require.Nil(t, integrationDependencyEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixIntegrationDependencyEntity(integrationDependencyID, appID)
		conv := integrationdependency.NewConverter(version.NewConverter(), aspect.NewConverter())

		integrationDependencyModel := conv.FromEntity(entity)

		assert.Equal(t, fixIntegrationDependencyModel(integrationDependencyID), integrationDependencyModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := integrationdependency.NewConverter(version.NewConverter(), aspect.NewConverter())

		integrationDependencyModel := conv.FromEntity(nil)

		require.Nil(t, integrationDependencyModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	modelIntDep := fixIntegrationDependencyModel(integrationDependencyID)
	gqlIntDep := fixGQLIntegrationDependency(integrationDependencyID)

	testCases := []struct {
		Name     string
		Input    *model.IntegrationDependency
		Expected *graphql.IntegrationDependency
	}{
		{
			Name:     "All properties given",
			Input:    modelIntDep,
			Expected: gqlIntDep,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVE
			converter := integrationdependency.NewConverter(version.NewConverter(), aspect.NewConverter())

			// WHEN
			res, err := converter.ToGraphQL(testCase.Input, []*model.Aspect{})

			// THEN
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	gqlIntegrationDependencyInput := fixGQLIntegrationDependencyInput()
	modelIntegrationDependencyInput := model.IntegrationDependencyInput{
		OrdID:         str.Ptr(ordID),
		Title:         title,
		Description:   str.Ptr(description),
		OrdPackageID:  str.Ptr(packageID),
		Visibility:    publicVisibility,
		ReleaseStatus: str.Ptr(releaseStatus),
		Mandatory:     &mandatory,
		Aspects: []*model.AspectInput{
			{
				Title:          title,
				Description:    str.Ptr(description),
				Mandatory:      &mandatory,
				APIResources:   json.RawMessage("[]"),
				EventResources: json.RawMessage("[]"),
			},
		},
	}

	testCases := []struct {
		Name     string
		Input    *graphql.IntegrationDependencyInput
		Expected *model.IntegrationDependencyInput
	}{
		{
			Name:     "All properties given",
			Input:    gqlIntegrationDependencyInput,
			Expected: &modelIntegrationDependencyInput,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVE
			converter := integrationdependency.NewConverter(version.NewConverter(), aspect.NewConverter())

			// WHEN
			res, err := converter.InputFromGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
