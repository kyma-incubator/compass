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
	testCases := []struct {
		Name              string
		Input             *model.Formation
		ExpectedFormation *graphql.Formation
		ExpectedError     error
	}{
		{
			Name:              "Success",
			Input:             &model.Formation{Name: testFormationName},
			ExpectedFormation: &graphql.Formation{Name: testFormationName},
		},
		{
			Name:              "Success when input is empty",
			Input:             nil,
			ExpectedFormation: nil,
		},
		{
			Name:              "Returns error when can't unmarshal the error",
			Input:             &model.Formation{Name: testFormationName, Error: json.RawMessage(`{invalid}`)},
			ExpectedFormation: nil,
			ExpectedError:     errors.New("while unmarshalling formation error"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result, err := converter.ToGraphQL(testCase.Input)

			if testCase.ExpectedError == nil {
				require.Equal(t, result, testCase.ExpectedFormation)
				require.NoError(t, err)
			} else {
				require.Nil(t, result)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
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
		ExpectedError      error
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
			ExpectedError:      errors.New("while unmarshalling formation error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result, err := converter.MultipleToGraphQL(testCase.InputFormations)

			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.ElementsMatch(t, testCase.ExpectedFormations, result)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
				require.Nil(t, result)
			}
		})
	}
}
