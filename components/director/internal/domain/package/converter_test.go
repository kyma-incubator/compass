package mp_package_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		name := "foo"
		desc := "bar"
		pkgModel := fixPackageModel(t, name, desc)
		require.NotNil(t, pkgModel)
		authConv := auth.NewConverter()
		conv := mp_package.NewConverter(authConv)
		//WHEN
		entity, err := conv.ToEntity(pkgModel)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, fixEntityPackage(packageID, name, desc), entity)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		name := "foo"
		pkgModel := &model.Package{
			ID:                             packageID,
			TenantID:                       tenantID,
			ApplicationID:                  appID,
			Name:                           name,
			Description:                    nil,
			InstanceAuthRequestInputSchema: nil,
			DefaultInstanceAuth:            nil,
		}

		expectedEntity := &mp_package.Entity{
			ID:                            packageID,
			TenantID:                      tenantID,
			ApplicationID:                 appID,
			Name:                          name,
			Description:                   sql.NullString{},
			InstanceAuthRequestJSONSchema: sql.NullString{},
			DefaultInstanceAuth:           sql.NullString{},
		}

		require.NotNil(t, pkgModel)
		authConv := auth.NewConverter()
		conv := mp_package.NewConverter(authConv)
		//WHEN
		entity, err := conv.ToEntity(pkgModel)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedEntity, entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		name := "foo"
		desc := "bar"
		entity := fixEntityPackage(packageID, name, desc)
		authConv := auth.NewConverter()
		conv := mp_package.NewConverter(authConv)
		//WHEN
		pkgModel, err := conv.FromEntity(entity)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, fixPackageModel(t, name, desc), pkgModel)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		name := "foo"
		entity := &mp_package.Entity{
			ID:                            packageID,
			TenantID:                      tenantID,
			ApplicationID:                 appID,
			Name:                          name,
			Description:                   sql.NullString{},
			InstanceAuthRequestJSONSchema: sql.NullString{},
			DefaultInstanceAuth:           sql.NullString{},
		}
		emptyString := ""
		expectedModel := &model.Package{
			ID:                             packageID,
			TenantID:                       tenantID,
			ApplicationID:                  appID,
			Name:                           name,
			Description:                    &emptyString,
			InstanceAuthRequestInputSchema: nil,
			DefaultInstanceAuth:            nil,
		}
		authConv := auth.NewConverter()
		conv := mp_package.NewConverter(authConv)
		//WHEN
		pkgModel, err := conv.FromEntity(entity)
		//THEN
		require.NoError(t, err)
		require.NotNil(t, expectedModel)
		assert.Equal(t, expectedModel, pkgModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	id := packageID
	name := "foo"
	desc := "bar"
	modelPackage := fixPackageModel(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)
	emptyModelPackage := &model.Package{}
	emptyGraphQLPackage := &graphql.Package{}

	testCases := []struct {
		Name            string
		Input           *model.Package
		Expected        *graphql.Package
		AuthConverterFn func() *automock.AuthConverter
		ExpectedErr     error
	}{
		{
			Name:     "All properties given",
			Input:    modelPackage,
			Expected: gqlPackage,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelPackage.DefaultInstanceAuth).Return(gqlPackage.DefaultInstanceAuth).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    emptyModelPackage,
			Expected: emptyGraphQLPackage,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", emptyModelPackage.DefaultInstanceAuth).Return(nil).Once()
				return conv
			},
		},
		{
			Name:        "Nil",
			Input:       nil,
			Expected:    nil,
			ExpectedErr: errors.New("the model Package is nil"),
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			authConverter := testCase.AuthConverterFn()

			// when
			converter := mp_package.NewConverter(authConverter)
			res, err := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			authConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	name1 := "foo"
	name2 := "bar"
	desc := "1"
	input := []*model.Package{
		fixPackageModel(t, name1, desc),
		fixPackageModel(t, name2, desc),
		{},
		nil,
	}

	expected := []*graphql.Package{
		fixGQLPackage(packageID, name1, desc),
		fixGQLPackage(packageID, name2, desc),
		{},
	}

	authConverter := &automock.AuthConverter{}

	for i, api := range input {
		if api == nil {
			continue
		}
		authConverter.On("ToGraphQL", api.DefaultInstanceAuth).Return(expected[i].DefaultInstanceAuth).Once()
	}

	// when
	converter := mp_package.NewConverter(authConverter)
	res, err := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	assert.NoError(t, err)

	authConverter.AssertExpectations(t)
}

func TestConverter_CreateInputFromGraphQL(t *testing.T) {
	// given
	name := "foo"
	desc := "Lorem ipsum"
	gqlPackageCreateInput := fixGQLPackageCreateInput(name, desc)
	modelPackageCreateInput := fixModelPackageCreateInput(t, name, desc)
	emptyGQLPackageCreateInput := &graphql.PackageCreateInput{}
	emptyModelPackageCreateInput := &model.PackageCreateInput{}
	testCases := []struct {
		Name            string
		Input           graphql.PackageCreateInput
		Expected        model.PackageCreateInput
		AuthConverterFn func() *automock.AuthConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlPackageCreateInput,
			Expected: modelPackageCreateInput,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlPackageCreateInput.DefaultInstanceAuth).Return(modelPackageCreateInput.DefaultInstanceAuth).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    graphql.PackageCreateInput{},
			Expected: model.PackageCreateInput{},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", emptyGQLPackageCreateInput.DefaultInstanceAuth).Return(emptyModelPackageCreateInput.DefaultInstanceAuth).Once()
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			authConverter := testCase.AuthConverterFn()

			// when
			converter := mp_package.NewConverter(authConverter)
			res, err := converter.CreateInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			assert.NoError(t, err)
			authConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_UpdateInputFromGraphQL(t *testing.T) {
	// given
	name := "foo"
	desc := "Lorem ipsum"
	gqlPackageCreateInput := fixGQLPackageUpdateInput(name, desc)
	modelPackageCreateInput := fixModelPackageUpdateInput(t, name, desc)
	emptyGQLPackageCreateInput := &graphql.PackageCreateInput{}
	emptyModelPackageCreateInput := &model.PackageCreateInput{}
	testCases := []struct {
		Name            string
		Input           graphql.PackageUpdateInput
		Expected        model.PackageUpdateInput
		AuthConverterFn func() *automock.AuthConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlPackageCreateInput,
			Expected: modelPackageCreateInput,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlPackageCreateInput.DefaultInstanceAuth).Return(modelPackageCreateInput.DefaultInstanceAuth).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    graphql.PackageUpdateInput{},
			Expected: model.PackageUpdateInput{},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", emptyGQLPackageCreateInput.DefaultInstanceAuth).Return(emptyModelPackageCreateInput.DefaultInstanceAuth).Once()
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			authConverter := testCase.AuthConverterFn()

			// when
			converter := mp_package.NewConverter(authConverter)
			res, err := converter.UpdateInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			assert.NoError(t, err)
			authConverter.AssertExpectations(t)
		})
	}
}
