package version_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Version
		Expected *graphql.Version
	}{
		{
			Name:     "All properties given",
			Input:    fixModelVersion(t, "foo", true, "bar", false),
			Expected: fixGQLVersion(t, "foo", true, "bar", false),
		},
		{
			Name:     "Empty",
			Input:    &model.Version{},
			Expected: &graphql.Version{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := version.NewConverter()

			// when
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.VersionInput
		Expected *model.VersionInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLVersionInput("foo", true, "bar", false),
			Expected: fixModelVersionInput("foo", true, "bar", false),
		},
		{
			Name:     "Empty",
			Input:    &graphql.VersionInput{},
			Expected: &model.VersionInput{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := version.NewConverter()

			// when
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
