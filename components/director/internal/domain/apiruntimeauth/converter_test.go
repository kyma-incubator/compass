package apiruntimeauth_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	rtmID := "foo"
	apiID := "bar"
	apiRtmAuthID := "baz"

	modelAuth := fixModelAuth()
	gqlAuth := fixGQLAuth()
	modelAPIRtmAuth := fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, modelAuth)
	gqlAPIRtmAuth := fixGQLAPIRuntimeAuth(rtmID, fixGQLAuth())

	testCases := []struct {
		Name           string
		AuthConvFn     func() *automock.AuthConverter
		Input          *model.APIRuntimeAuth
		ExpectedOutput *graphql.APIRuntimeAuth
	}{
		{
			Name: "Success",
			AuthConvFn: func() *automock.AuthConverter {
				authConv := &automock.AuthConverter{}
				authConv.On("ToGraphQL", modelAuth).Return(gqlAuth).Once()
				return authConv
			},
			Input:          modelAPIRtmAuth,
			ExpectedOutput: gqlAPIRtmAuth,
		},
		{
			Name: "Returns nil when input is nil",
			AuthConvFn: func() *automock.AuthConverter {
				authConv := &automock.AuthConverter{}
				return authConv
			},
			Input:          nil,
			ExpectedOutput: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := testCase.AuthConvFn()
			conv := apiruntimeauth.NewConverter(authConv)

			// WHEN
			result := conv.ToGraphQL(testCase.Input)

			// THEN
			assert.Equal(t, testCase.ExpectedOutput, result)

			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	// GIVEN
	rtmID := "foo"
	apiID := "bar"
	apiRtmAuthID := "baz"

	modelAPIRtmAuth := *fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, fixModelAuth())
	modelAPIRtmAuthWithNils := *fixModelAPIRuntimeAuth(nil, rtmID, apiID, nil)
	ent := fixEntity(&apiRtmAuthID, rtmID, apiID, true)
	entWithNils := fixEntity(nil, rtmID, apiID, false)

	testCases := []struct {
		Name           string
		Input          model.APIRuntimeAuth
		ExpectedOutput apiruntimeauth.Entity
		ExpectedError  error
	}{
		{
			Name:           "Success",
			Input:          modelAPIRtmAuth,
			ExpectedOutput: ent,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when optional fields are nil",
			Input:          modelAPIRtmAuthWithNils,
			ExpectedOutput: entWithNils,
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apiruntimeauth.NewConverter(nil)

			// WHEN
			result, err := conv.ToEntity(testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// GIVEN
	rtmID := "foo"
	apiID := "bar"
	apiRtmAuthID := "baz"

	modelAPIRtmAuth := *fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, fixModelAuth())
	modelAPIRtmAuthWithNils := *fixModelAPIRuntimeAuth(nil, rtmID, apiID, nil)
	ent := fixEntity(&apiRtmAuthID, rtmID, apiID, true)
	entWithNils := fixEntity(nil, rtmID, apiID, false)

	testCases := []struct {
		Name           string
		Input          apiruntimeauth.Entity
		ExpectedOutput model.APIRuntimeAuth
		ExpectedError  error
	}{
		{
			Name:           "Success",
			Input:          ent,
			ExpectedOutput: modelAPIRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when optional fields are nil",
			Input:          entWithNils,
			ExpectedOutput: modelAPIRtmAuthWithNils,
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apiruntimeauth.NewConverter(nil)

			// WHEN
			result, err := conv.FromEntity(testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}
