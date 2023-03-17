package formation_test

import (
	"encoding/json"
	"testing"

	"github.com/pkg/errors"

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
		actual, err := converter.ToGraphQL(&model.Formation{Name: testFormationName})

		// THEN
		require.NoError(t, err)
		require.Equal(t, testFormationName, actual.Name)
	})

	t.Run("Success when input is empty", func(t *testing.T) {
		// WHEN
		actual, err := converter.ToGraphQL(nil)

		// THEN
		require.NoError(t, err)
		require.Nil(t, actual)
	})

	t.Run("Returns error when can't unmarshal the error", func(t *testing.T) {
		// WHEN
		actual, err := converter.ToGraphQL(&model.Formation{
			Name:  testFormationName,
			Error: json.RawMessage(`{invalid}`),
		})

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "while unmarshalling formation error")
		require.Nil(t, actual)
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
		ExpectedErrorMsg   error
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
		{
			Name:               "Returns error when can't convert one of input formations",
			InputFormations:    []*model.Formation{&modelFormation, {Error: json.RawMessage(`{invalid}`)}},
			ExpectedFormations: nil,
			ExpectedErrorMsg:   errors.New("while unmarshalling formation error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result, err := converter.MultipleToGraphQL(testCase.InputFormations)

			if testCase.ExpectedErrorMsg == nil {
				require.NoError(t, err)
				require.ElementsMatch(t, testCase.ExpectedFormations, result)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg.Error())
				require.Nil(t, result)
			}
		})
	}
}
