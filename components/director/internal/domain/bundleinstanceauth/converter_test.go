package bundleinstanceauth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testJSON = graphql.JSON("test")
	testStr  = "test"

	authInputModel = fixModelAuthInput()
	authInputGQL   = fixGQLAuthInput()
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	authModel := fixModelAuth()
	authGQL := fixGQLAuth()

	piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, authModel, fixModelStatusSucceeded(), &testRuntimeID)
	piaGQL := fixGQLBundleInstanceAuth(testID, authGQL, fixGQLStatusSucceeded(), &testRuntimeID)

	testCases := []struct {
		Name            string
		AuthConverterFn func() *automock.AuthConverter
		Input           *model.BundleInstanceAuth
		Expected        *graphql.BundleInstanceAuth
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
				conv.On("ToGraphQL", piaModel.Auth).Return(piaGQL.Auth, nil).Once()
				return conv
			},
			Input:    piaModel,
			Expected: piaGQL,
		},
		{
			Name: "Success when context and input params empty",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", (*model.Auth)(nil)).Return(nil, nil).Once()
				return conv
			},
			Input:    fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil),
			Expected: fixGQLBundleInstanceAuthWithoutContextAndInputParams(testID, nil, fixGQLStatusPending(), nil),
		},
		{
			Name: "Success when context and input params empty",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", (*model.Auth)(nil)).Return(nil, nil).Once()
				return conv
			},
			Input:    fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, nil),
			Expected: fixGQLBundleInstanceAuthWithoutContextAndInputParams(testID, nil, nil, nil),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := bundleinstanceauth.NewConverter(authConv)
			// WHEN
			result, err := conv.ToGraphQL(testCase.Input)

			// THEN
			require.NoError(t, err)
			require.Equal(t, testCase.Expected, result)

			mock.AssertExpectationsForObjects(t, authConv)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	piaModels := []*model.BundleInstanceAuth{
		fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
		fixModelBundleInstanceAuth("bar", testBundleID, testTenant, nil, fixModelStatusPending(), &testRuntimeID),
		nil,
	}

	piaGQLs := []*graphql.BundleInstanceAuth{
		fixGQLBundleInstanceAuth("foo", fixGQLAuth(), fixGQLStatusSucceeded(), nil),
		fixGQLBundleInstanceAuth("bar", nil, fixGQLStatusPending(), &testRuntimeID),
	}

	testCases := []struct {
		Name            string
		AuthConverterFn func() *automock.AuthConverter
		Input           []*model.BundleInstanceAuth
		Expected        []*graphql.BundleInstanceAuth
	}{
		{
			Name: "Success when nil",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			Input:    nil,
			Expected: []*graphql.BundleInstanceAuth{},
		},
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", piaModels[0].Auth).Return(piaGQLs[0].Auth, nil).Once()
				conv.On("ToGraphQL", piaModels[1].Auth).Return(piaGQLs[1].Auth, nil).Once()
				return conv
			},
			Input:    piaModels,
			Expected: piaGQLs,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := bundleinstanceauth.NewConverter(authConv)
			// WHEN
			result, err := conv.MultipleToGraphQL(testCase.Input)

			// THEN
			require.NoError(t, err)
			require.Equal(t, testCase.Expected, result)

			mock.AssertExpectationsForObjects(t, authConv)
		})
	}
}

func TestConverter_RequestInputFromGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    graphql.BundleInstanceAuthRequestInput
		Expected model.BundleInstanceAuthRequestInput
	}{
		{
			Name: "Success when nil",
			Input: graphql.BundleInstanceAuthRequestInput{
				Context:     nil,
				InputParams: nil,
			},
			Expected: model.BundleInstanceAuthRequestInput{
				Context:     nil,
				InputParams: nil,
			},
		},
		{
			Name: "Success when not nil",
			Input: graphql.BundleInstanceAuthRequestInput{
				Context:     &testJSON,
				InputParams: &testJSON,
			},
			Expected: model.BundleInstanceAuthRequestInput{
				Context:     &testStr,
				InputParams: &testStr,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := bundleinstanceauth.NewConverter(nil)

			// WHEN
			result := conv.RequestInputFromGraphQL(testCase.Input)

			// THEN
			require.Equal(t, testCase.Expected, result)
		})
	}
}

func TestConverter_CreateInputFromGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name             string
		AuthConverterFn  func() *automock.AuthConverter
		Input            graphql.BundleInstanceAuthCreateInput
		ExpectedOutput   model.BundleInstanceAuthCreateInput
		ExpectedErrorMsg string
	}{
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(authInputModel, nil).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthCreateInput{
				Context:     &testJSON,
				InputParams: &testJSON,
				Auth:        authInputGQL,
				RuntimeID:   &testStr,
			},
			ExpectedOutput: model.BundleInstanceAuthCreateInput{
				Context:     &testStr,
				InputParams: &testStr,
				Auth:        authInputModel,
				RuntimeID:   &testStr,
			},
		},
		{
			Name: "Returns error when can't convert auth",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(nil, testError).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthCreateInput{
				Context:     &testJSON,
				InputParams: &testJSON,
				Auth:        authInputGQL,
				RuntimeID:   &testStr,
			},
			ExpectedErrorMsg: "while converting Auth",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := bundleinstanceauth.NewConverter(authConv)

			// WHEN
			result, err := conv.CreateInputFromGraphQL(testCase.Input)

			// THEN
			if testCase.ExpectedErrorMsg == "" {
				require.Equal(t, testCase.ExpectedOutput, result)
				require.Nil(t, err)
			} else {
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			}

			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_UpdateInputFromGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name             string
		AuthConverterFn  func() *automock.AuthConverter
		Input            graphql.BundleInstanceAuthUpdateInput
		ExpectedOutput   model.BundleInstanceAuthUpdateInput
		ExpectedErrorMsg string
	}{
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(authInputModel, nil).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthUpdateInput{
				Context:     &testJSON,
				InputParams: &testJSON,
				Auth:        authInputGQL,
			},
			ExpectedOutput: model.BundleInstanceAuthUpdateInput{
				Context:     &testStr,
				InputParams: &testStr,
				Auth:        authInputModel,
			},
		},
		{
			Name: "Returns error when can't convert auth",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(nil, testError).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthUpdateInput{
				Context:     &testJSON,
				InputParams: &testJSON,
				Auth:        authInputGQL,
			},
			ExpectedErrorMsg: "while converting Auth",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := bundleinstanceauth.NewConverter(authConv)

			// WHEN
			result, err := conv.UpdateInputFromGraphQL(testCase.Input)

			// THEN
			if testCase.ExpectedErrorMsg == "" {
				require.Equal(t, testCase.ExpectedOutput, result)
				require.Nil(t, err)
			} else {
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			}

			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_SetInputFromGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name            string
		AuthConverterFn func() *automock.AuthConverter
		Input           graphql.BundleInstanceAuthSetInput
		Expected        model.BundleInstanceAuthSetInput
	}{
		{
			Name: "Success",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(authInputModel, nil).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthSetInput{
				Auth:   authInputGQL,
				Status: fixGQLStatusInput(graphql.BundleInstanceAuthSetStatusConditionInputSucceeded, "foo", "bar"),
			},
			Expected: model.BundleInstanceAuthSetInput{
				Auth:   authInputModel,
				Status: fixModelStatusInput(model.BundleInstanceAuthSetStatusConditionInputSucceeded, "foo", "bar"),
			},
		},
		{
			Name: "Success when no status",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", authInputGQL).Return(authInputModel, nil).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthSetInput{
				Auth:   authInputGQL,
				Status: nil,
			},
			Expected: model.BundleInstanceAuthSetInput{
				Auth:   authInputModel,
				Status: nil,
			},
		},
		{
			Name: "Success when no auth",
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", (*graphql.AuthInput)(nil)).Return(nil, nil).Once()
				return conv
			},
			Input: graphql.BundleInstanceAuthSetInput{
				Auth:   nil,
				Status: fixGQLStatusInput(graphql.BundleInstanceAuthSetStatusConditionInputFailed, "foo", "bar"),
			},
			Expected: model.BundleInstanceAuthSetInput{
				Auth:   nil,
				Status: fixModelStatusInput(model.BundleInstanceAuthSetStatusConditionInputFailed, "foo", "bar"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConverterFn()

			conv := bundleinstanceauth.NewConverter(authConv)
			// WHEN
			result, err := conv.SetInputFromGraphQL(testCase.Input)

			// THEN
			require.NoError(t, err)
			require.Equal(t, testCase.Expected, result)

			mock.AssertExpectationsForObjects(t, authConv)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	t.Run("Success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID)
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID)

		conv := bundleinstanceauth.NewConverter(nil)

		// WHEN
		entity, err := conv.ToEntity(piaModel)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, piaEntity, entity)
	})

	t.Run("Success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		piaModel := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, nil, nil)
		piaEntity := fixEntityBundleInstanceAuthWithoutContextAndInputParams(t, testID, testBundleID, testTenant, nil, nil, nil)

		conv := bundleinstanceauth.NewConverter(nil)

		// WHEN
		entity, err := conv.ToEntity(piaModel)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, piaEntity, entity)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	t.Run("Success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID)
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID)

		conv := bundleinstanceauth.NewConverter(nil)

		// WHEN
		result, err := conv.FromEntity(piaEntity)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, piaModel, result)
	})

	t.Run("Success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		piaModel := fixModelBundleInstanceAuthWithoutContextAndInputParams(testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)
		piaEntity := fixEntityBundleInstanceAuthWithoutContextAndInputParams(t, testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)

		conv := bundleinstanceauth.NewConverter(nil)

		// WHEN
		result, err := conv.FromEntity(piaEntity)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, piaModel, result)
	})
}
