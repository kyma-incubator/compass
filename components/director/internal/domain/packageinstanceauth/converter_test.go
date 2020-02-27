package packageinstanceauth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	authModel := fixModelAuth()
	authGQL := fixGQLAuth()

	piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, authModel, fixModelStatusSucceeded())
	piaGQL := fixGQLPackageInstanceAuth(testID, authGQL, fixGQLStatusSucceeded())

	testCases := []struct {
		Name            string
		AuthConverterFn func() *automock.AuthConverter
		Input           *model.PackageInstanceAuth
		Expected        *graphql.PackageInstanceAuth
	}{
		{
			Name: "Success when nil",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			Input:    nil,
			Expected: nil,
		},
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", piaModel.Auth).Return(piaGQL.Auth).Once()
				return conv
			},
			Input:    piaModel,
			Expected: piaGQL,
		},
		{
			Name: "Success when context and input params empty",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", (*model.Auth)(nil)).Return(nil).Once()
				return conv
			},
			Input:    fixModelPackageInstanceAuthWithoutContextAndInputParams(testID, testPackageID, testTenant, nil, fixModelStatusPending()),
			Expected: fixGQLPackageInstanceAuthWithoutContextAndInputParams(testID, nil, fixGQLStatusPending()),
		},
		{
			Name: "Success when context and input params empty",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", (*model.Auth)(nil)).Return(nil).Once()
				return conv
			},
			Input:    fixModelPackageInstanceAuthWithoutContextAndInputParams(testID, testPackageID, testTenant, nil, nil),
			Expected: fixGQLPackageInstanceAuthWithoutContextAndInputParams(testID, nil, nil),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := packageinstanceauth.NewConverter(authConv)
			// WHEN
			result := conv.ToGraphQL(testCase.Input)

			// THEN
			require.Equal(t, testCase.Expected, result)

			mock.AssertExpectationsForObjects(t, authConv)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	piaModels := []*model.PackageInstanceAuth{
		fixModelPackageInstanceAuth("foo", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		fixModelPackageInstanceAuth("bar", testPackageID, testTenant, nil, fixModelStatusPending()),
		nil,
	}

	piaGQLs := []*graphql.PackageInstanceAuth{
		fixGQLPackageInstanceAuth("foo", fixGQLAuth(), fixGQLStatusSucceeded()),
		fixGQLPackageInstanceAuth("bar", nil, fixGQLStatusPending()),
	}

	testCases := []struct {
		Name            string
		AuthConverterFn func() *automock.AuthConverter
		Input           []*model.PackageInstanceAuth
		Expected        []*graphql.PackageInstanceAuth
	}{
		{
			Name: "Success when nil",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			Input:    nil,
			Expected: nil,
		},
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", piaModels[0].Auth).Return(piaGQLs[0].Auth).Once()
				conv.On("ToGraphQL", piaModels[1].Auth).Return(piaGQLs[1].Auth).Once()
				return conv
			},
			Input:    piaModels,
			Expected: piaGQLs,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := packageinstanceauth.NewConverter(authConv)
			// WHEN
			result := conv.MultipleToGraphQL(testCase.Input)

			// THEN
			require.Equal(t, testCase.Expected, result)

			mock.AssertExpectationsForObjects(t, authConv)
		})
	}
}

func TestConverter_RequestInputFromGraphQL(t *testing.T) {
	// GIVEN
	testJSON := graphql.JSON("test")
	testStr := "test"

	testCases := []struct {
		Name     string
		Input    graphql.PackageInstanceAuthRequestInput
		Expected model.PackageInstanceAuthRequestInput
	}{
		{
			Name: "Success when nil",
			Input: graphql.PackageInstanceAuthRequestInput{
				Context:     nil,
				InputParams: nil,
			},
			Expected: model.PackageInstanceAuthRequestInput{
				Context:     nil,
				InputParams: nil,
			},
		},
		{
			Name: "Success when not nil",
			Input: graphql.PackageInstanceAuthRequestInput{
				Context:     &testJSON,
				InputParams: &testJSON,
			},
			Expected: model.PackageInstanceAuthRequestInput{
				Context:     &testStr,
				InputParams: &testStr,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := packageinstanceauth.NewConverter(nil)

			// WHEN
			result := conv.RequestInputFromGraphQL(testCase.Input)

			// THEN
			require.Equal(t, testCase.Expected, result)
		})
	}
}

func TestConverter_SetInputFromGraphQL(t *testing.T) {
	// GIVEN
	authInputModel := fixModelAuthInput()
	authInputGQL := fixGQLAuthInput()

	testCases := []struct {
		Name            string
		AuthConverterFn func() *automock.AuthConverter
		Input           graphql.PackageInstanceAuthSetInput
		Expected        model.PackageInstanceAuthSetInput
	}{
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(authInputModel).Once()
				return conv
			},
			Input: graphql.PackageInstanceAuthSetInput{
				Auth:   authInputGQL,
				Status: fixGQLStatusInput(graphql.PackageInstanceAuthSetStatusConditionInputSucceeded, "foo", "bar"),
			},
			Expected: model.PackageInstanceAuthSetInput{
				Auth:   authInputModel,
				Status: fixModelStatusInput(model.PackageInstanceAuthSetStatusConditionInputSucceeded, "foo", "bar"),
			},
		},
		{
			Name: "Success when no status",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(authInputModel).Once()
				return conv
			},
			Input: graphql.PackageInstanceAuthSetInput{
				Auth:   authInputGQL,
				Status: nil,
			},
			Expected: model.PackageInstanceAuthSetInput{
				Auth:   authInputModel,
				Status: nil,
			},
		},
		{
			Name: "Success when no auth",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", (*graphql.AuthInput)(nil)).Return(nil).Once()
				return conv
			},
			Input: graphql.PackageInstanceAuthSetInput{
				Auth:   nil,
				Status: fixGQLStatusInput(graphql.PackageInstanceAuthSetStatusConditionInputFailed, "foo", "bar"),
			},
			Expected: model.PackageInstanceAuthSetInput{
				Auth:   nil,
				Status: fixModelStatusInput(model.PackageInstanceAuthSetStatusConditionInputFailed, "foo", "bar"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := packageinstanceauth.NewConverter(authConv)
			// WHEN
			result := conv.SetInputFromGraphQL(testCase.Input)

			// THEN
			require.Equal(t, testCase.Expected, result)

			mock.AssertExpectationsForObjects(t, authConv)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	t.Run("Success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		conv := packageinstanceauth.NewConverter(nil)

		//WHEN
		entity, err := conv.ToEntity(*piaModel)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, piaEntity, &entity)
	})

	t.Run("Success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		piaModel := fixModelPackageInstanceAuthWithoutContextAndInputParams(testID, testPackageID, testTenant, nil, nil)
		piaEntity := fixEntityPackageInstanceAuthWithoutContextAndInputParams(t, testID, testPackageID, testTenant, nil, nil)

		conv := packageinstanceauth.NewConverter(nil)

		//WHEN
		entity, err := conv.ToEntity(*piaModel)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, piaEntity, &entity)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	t.Run("Success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		conv := packageinstanceauth.NewConverter(nil)

		//WHEN
		result, err := conv.FromEntity(*piaEntity)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, piaModel, &result)
	})

	t.Run("Success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		piaModel := fixModelPackageInstanceAuthWithoutContextAndInputParams(testID, testPackageID, testTenant, nil, fixModelStatusPending())
		piaEntity := fixEntityPackageInstanceAuthWithoutContextAndInputParams(t, testID, testPackageID, testTenant, nil, fixModelStatusPending())

		conv := packageinstanceauth.NewConverter(nil)

		//WHEN
		result, err := conv.FromEntity(*piaEntity)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, piaModel, &result)
	})
}
