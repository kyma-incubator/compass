package formation_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	t.Run("Success when state is provided externally", func(t *testing.T) {
		// WHEN
		actual := converter.FromGraphQL(graphql.FormationInput{Name: testFormationName, State: str.Ptr(string(model.InitialFormationState))})

		// THEN
		require.Equal(t, testFormationName, actual.Name)
		require.Equal(t, model.InitialFormationState, actual.State)
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

func Test_converter_MultipleToGraphQL(t *testing.T) {
	testCases := []struct {
		Name               string
		InputFormations    []*model.Formation
		ExpectedFormations []*graphql.Formation
	}{
		{
			Name:               "Success",
			InputFormations:    []*model.Formation{&modelFormation},
			ExpectedFormations: []*graphql.Formation{&graphqlFormation},
		},
		{
			Name:               "Success when input is nil",
			InputFormations:    nil,
			ExpectedFormations: nil,
		},
		{
			Name:               "Success when one of the input formations is nil",
			InputFormations:    []*model.Formation{nil, &modelFormation},
			ExpectedFormations: []*graphql.Formation{&graphqlFormation},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := converter.MultipleToGraphQL(testCase.InputFormations)

			require.ElementsMatch(t, testCase.ExpectedFormations, result)
		})
	}
}
