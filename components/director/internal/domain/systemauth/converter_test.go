package systemauth_test

import (
	"testing"

	systemauth2 "github.com/kyma-incubator/compass/components/director/pkg/systemauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	sysAuthID := "foo"
	objectID := "bar"

	modelAuth := fixModelAuth()
	gqlAuth := fixGQLAuth()
	modelRtmSysAuth := fixModelSystemAuth(sysAuthID, systemauth2.RuntimeReference, objectID, modelAuth)
	modelAppSysAuth := fixModelSystemAuth(sysAuthID, systemauth2.ApplicationReference, objectID, modelAuth)
	modelIntSysAuth := fixModelSystemAuth(sysAuthID, systemauth2.IntegrationSystemReference, objectID, modelAuth)

	testCases := []struct {
		Name           string
		AuthConvFn     func() *automock.AuthConverter
		Input          *systemauth2.SystemAuth
		ExpectedOutput graphql.SystemAuth
	}{
		{
			Name: "Success when converting auth for Runtime",
			AuthConvFn: func() *automock.AuthConverter {
				authConv := &automock.AuthConverter{}
				authConv.On("ToGraphQL", modelAuth).Return(gqlAuth, nil).Once()
				return authConv
			},
			Input:          modelRtmSysAuth,
			ExpectedOutput: fixGQLRuntimeSystemAuth(sysAuthID, gqlAuth),
		},
		{
			Name: "Success when converting auth for Application",
			AuthConvFn: func() *automock.AuthConverter {
				authConv := &automock.AuthConverter{}
				authConv.On("ToGraphQL", modelAuth).Return(gqlAuth, nil).Once()
				return authConv
			},
			Input:          modelAppSysAuth,
			ExpectedOutput: fixGQLAppSystemAuth(sysAuthID, gqlAuth),
		},
		{
			Name: "Success when converting auth for Integration System",
			AuthConvFn: func() *automock.AuthConverter {
				authConv := &automock.AuthConverter{}
				authConv.On("ToGraphQL", modelAuth).Return(gqlAuth, nil).Once()
				return authConv
			},
			Input:          modelIntSysAuth,
			ExpectedOutput: fixGQLIntSysSystemAuth(sysAuthID, gqlAuth),
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
			conv := systemauth.NewConverter(authConv)

			// WHEN
			result, err := conv.ToGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, testCase.ExpectedOutput, result)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	// GIVEN
	sysAuthID := "foo"
	objectID := "bar"

	modelAuth := fixModelAuth()
	modelRtmSysAuth := *fixModelSystemAuth(sysAuthID, systemauth2.RuntimeReference, objectID, modelAuth)
	modelAppSysAuth := *fixModelSystemAuth(sysAuthID, systemauth2.ApplicationReference, objectID, modelAuth)
	modelIntSysAuth := *fixModelSystemAuth(sysAuthID, systemauth2.IntegrationSystemReference, objectID, modelAuth)

	entRtm := fixEntity(sysAuthID, systemauth2.RuntimeReference, objectID, true)
	entApp := fixEntity(sysAuthID, systemauth2.ApplicationReference, objectID, true)
	entInt := fixEntity(sysAuthID, systemauth2.IntegrationSystemReference, objectID, true)

	testCases := []struct {
		Name           string
		Input          systemauth2.SystemAuth
		ExpectedOutput systemauth.Entity
		ExpectedError  error
	}{
		{
			Name:           "Success when converting auth for Runtime",
			Input:          modelRtmSysAuth,
			ExpectedOutput: entRtm,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when converting auth for Application",
			Input:          modelAppSysAuth,
			ExpectedOutput: entApp,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when converting auth for Integration System",
			Input:          modelIntSysAuth,
			ExpectedOutput: entInt,
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := systemauth.NewConverter(nil)

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
	sysAuthID := "foo"
	objectID := "bar"

	modelAuth := fixModelAuth()
	modelRtmSysAuth := *fixModelSystemAuth(sysAuthID, systemauth2.RuntimeReference, objectID, modelAuth)
	modelAppSysAuth := *fixModelSystemAuth(sysAuthID, systemauth2.ApplicationReference, objectID, modelAuth)
	modelIntSysAuth := *fixModelSystemAuth(sysAuthID, systemauth2.IntegrationSystemReference, objectID, modelAuth)

	entRtm := fixEntity(sysAuthID, systemauth2.RuntimeReference, objectID, true)
	entApp := fixEntity(sysAuthID, systemauth2.ApplicationReference, objectID, true)
	entInt := fixEntity(sysAuthID, systemauth2.IntegrationSystemReference, objectID, true)

	testCases := []struct {
		Name           string
		Input          systemauth.Entity
		ExpectedOutput systemauth2.SystemAuth
		ExpectedError  error
	}{
		{
			Name:           "Success when converting auth for Runtime",
			Input:          entRtm,
			ExpectedOutput: modelRtmSysAuth,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when converting auth for Application",
			Input:          entApp,
			ExpectedOutput: modelAppSysAuth,
			ExpectedError:  nil,
		},
		{
			Name:           "Success when converting auth for Integration System",
			Input:          entInt,
			ExpectedOutput: modelIntSysAuth,
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := systemauth.NewConverter(nil)

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
