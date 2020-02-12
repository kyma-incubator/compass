package packageinstanceauth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

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

func TestConverter_StatusToGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.PackageInstanceAuthStatus
		Expected *graphql.PackageInstanceAuthStatus
	}{
		{
			Name:     "Success when nil",
			Input:    nil,
			Expected: nil,
		},
		{
			Name:     "Success when status is Succeeded",
			Input:    fixModelStatusSucceeded(),
			Expected: fixGQLStatusSucceeded(),
		},
		{
			Name:     "Success when status is Pending",
			Input:    fixModelStatusPending(),
			Expected: fixGQLStatusPending(),
		},
		{
			Name: "Success when no message and reason",
			Input: &model.PackageInstanceAuthStatus{
				Condition: model.PackageInstanceAuthStatusConditionSucceeded,
				Timestamp: testTime,
				Message:   nil,
				Reason:    nil,
			},
			Expected: &graphql.PackageInstanceAuthStatus{
				Condition: graphql.PackageInstanceAuthStatusConditionSucceeded,
				Timestamp: graphql.Timestamp(testTime),
				Message:   "",
				Reason:    "",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := packageinstanceauth.NewConverter(nil)
			// WHEN
			result := conv.StatusToGraphQL(testCase.Input)

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
				Status: fixGQLStatusInput(graphql.PackageInstanceAuthSetStatusConditionInputSucceeded, str.Ptr("foo"), str.Ptr("bar")),
			},
			Expected: model.PackageInstanceAuthSetInput{
				Auth:   authInputModel,
				Status: fixModelStatusInput(model.PackageInstanceAuthSetStatusConditionInputSucceeded, str.Ptr("foo"), str.Ptr("bar")),
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
				Status: fixGQLStatusInput(graphql.PackageInstanceAuthSetStatusConditionInputFailed, str.Ptr("foo"), str.Ptr("bar")),
			},
			Expected: model.PackageInstanceAuthSetInput{
				Auth:   nil,
				Status: fixModelStatusInput(model.PackageInstanceAuthSetStatusConditionInputFailed, str.Ptr("foo"), str.Ptr("bar")),
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
