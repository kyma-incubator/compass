package runtime_auth_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_auth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	rtmID := "foo"
	apiID := "bar"
	rtmAuthID := "baz"

	modelAuth := fixModelAuth()
	gqlAuth := fixGQLAuth()
	modelRtmAuth := fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, modelAuth)
	gqlRtmAuth := fixGQLRuntimeAuth(rtmID, fixGQLAuth())

	testCases := []struct {
		Name           string
		AuthConvFn     func() *automock.AuthConverter
		Input          *model.RuntimeAuth
		ExpectedOutput *graphql.RuntimeAuth
	}{
		{
			Name: "Success",
			AuthConvFn: func() *automock.AuthConverter {
				authConv := &automock.AuthConverter{}
				authConv.On("ToGraphQL", modelAuth).Return(gqlAuth).Once()
				return authConv
			},
			Input:          modelRtmAuth,
			ExpectedOutput: gqlRtmAuth,
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
			conv := runtime_auth.NewConverter(authConv)

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
	rtmAuthID := "baz"

	modelRtmAuth := *fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, fixModelAuth())
	modelRtmAuthWithNils := *fixModelRuntimeAuth(nil, rtmID, apiID, nil)
	ent := fixEntity(&rtmAuthID, rtmID, apiID, true)
	entWithNils := fixEntity(nil, rtmID, apiID, false)

	testCases := []struct {
		Name           string
		Input          model.RuntimeAuth
		ExpectedOutput runtime_auth.Entity
		ExpectedError  error
	}{
		{
			Name:           "Success",
			Input:          modelRtmAuth,
			ExpectedOutput: ent,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when optional fields are nil",
			Input:          modelRtmAuthWithNils,
			ExpectedOutput: entWithNils,
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := runtime_auth.NewConverter(nil)

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
	rtmAuthID := "baz"

	modelRtmAuth := *fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, fixModelAuth())
	modelRtmAuthWithNils := *fixModelRuntimeAuth(nil, rtmID, apiID, nil)
	ent := fixEntity(&rtmAuthID, rtmID, apiID, true)
	entWithNils := fixEntity(nil, rtmID, apiID, false)

	testCases := []struct {
		Name           string
		Input          runtime_auth.Entity
		ExpectedOutput model.RuntimeAuth
		ExpectedError  error
	}{
		{
			Name:           "Success",
			Input:          ent,
			ExpectedOutput: modelRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when optional fields are nil",
			Input:          entWithNils,
			ExpectedOutput: modelRtmAuthWithNils,
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := runtime_auth.NewConverter(nil)

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
