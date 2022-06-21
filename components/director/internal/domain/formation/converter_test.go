package formation_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var converter = formation.NewConverter()

func TestFromGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.FromGraphQL(graphql.FormationInput{Name: testFormationName})

		// THEN
		require.Equal(t, testFormationName, actual.Name)
	})
}

func TestToGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.ToGraphQL(&model.Formation{Name: testFormationName})

		// THEN
		require.Equal(t, testFormationName, actual.Name)
	})
}

func TestConverter_ToEntity(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.Formation
		Expected *formation.Entity
	}{
		{
			Name:     "Success",
			Input:    fixFormationModel(),
			Expected: fixFormationEntity(),
		}, {
			Name:     "Returns nil when given empty model",
			Input:    nil,
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := converter.ToEntity(testCase.Input)

			require.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *formation.Entity
		Expected *model.Formation
	}{
		{
			Name:     "Success",
			Input:    fixFormationEntity(),
			Expected: fixFormationModel(),
		}, {
			Name:     "Empty",
			Input:    nil,
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := converter.FromEntity(testCase.Input)

			require.Equal(t, result, testCase.Expected)
		})
	}
}
