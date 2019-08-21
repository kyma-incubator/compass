package version_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/require"

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
			Input:    fixModelVersion("foo", true, "bar", false),
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

func TestConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		versionEntity := fixVersionEntity("v1.2", true, "v1.2", false)
		versionConv := version.NewConverter()
		//WHEN
		versionModel, err := versionConv.FromEntity(versionEntity)
		//THEN
		require.NoError(t, err)
		assertVersion(t, versionEntity, versionModel)
	})

	t.Run("success all nullable properties empty", func(t *testing.T) {
		versionEntity := version.Version{VersionValue: "v1"}
		versionConv := version.NewConverter()
		// WHEN
		versionModel, err := versionConv.FromEntity(&versionEntity)
		//THEN
		require.NoError(t, err)
		assertVersion(t, &versionEntity, versionModel)

	})
}
func TestConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		versionModel := fixModelVersion("v1.2", true, "v1.2", false)
		versionConv := version.NewConverter()
		//WHEN
		versionEntity, err := versionConv.ToEntity(versionModel)
		//THEN
		require.NoError(t, err)
		assertVersion(t, versionEntity, versionModel)
	})

	t.Run("success all nullable properties empty", func(t *testing.T) {
		versionModel := &model.Version{Value: "v1"}
		versionConv := version.NewConverter()
		//WHEN
		versionEntity, err := versionConv.ToEntity(versionModel)
		//THEN
		require.NoError(t, err)
		assertVersion(t, versionEntity, versionModel)
	})
}

func assertVersion(t *testing.T, entity *version.Version, model *model.Version) {
	assert.Equal(t, model.Value, entity.VersionValue)
	repo.AssertSqlNullString(t, &entity.VersionDepracatedSince, model.DeprecatedSince)
	repo.AssertSqlNullBool(t, &entity.VersionDepracated, model.Deprecated)
	repo.AssertSqlNullBool(t, &entity.VersionForRemoval, model.ForRemoval)
}
