package oauth20_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_CommonRequestClientCredentialsSuccess(t *testing.T) {
	// GIVEN
	id := "foo"
	clientID := "clientid"
	txGen := txtest.NewTransactionContextGenerator(nil)
	expectedResult := fixGQLSystemAuth(clientID)
	credsData := &model.OAuthCredentialDataInput{
		ClientID:     "clientid",
		ClientSecret: "secret",
		URL:          "url",
	}
	authInput := &model.AuthInput{Credential: &model.CredentialDataInput{Oauth: credsData}}

	testCases := []struct {
		Name                       string
		ObjType                    pkgmodel.SystemAuthReferenceObjectType
		Method                     func(resolver *oauth20.Resolver, ctx context.Context, id string) (graphql.SystemAuth, error)
		RtmID                      *string
		AppID                      *string
		IntSysID                   *string
		RuntimeServiceFn           func() *automock.RuntimeService
		ApplicationServiceFn       func() *automock.ApplicationService
		IntegrationSystemServiceFn func() *automock.IntegrationSystemService
	}{
		{
			Name:    "Runtime",
			RtmID:   &id,
			ObjType: pkgmodel.RuntimeReference,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			IntegrationSystemServiceFn: func() *automock.IntegrationSystemService {
				isSvc := &automock.IntegrationSystemService{}
				return isSvc
			},
			Method: func(resolver *oauth20.Resolver, ctx context.Context, id string) (graphql.SystemAuth, error) {
				return resolver.RequestClientCredentialsForRuntime(ctx, id)
			},
		},
		{
			Name:    "Application",
			AppID:   &id,
			ObjType: pkgmodel.ApplicationReference,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				return rtmSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return appSvc
			},
			IntegrationSystemServiceFn: func() *automock.IntegrationSystemService {
				isSvc := &automock.IntegrationSystemService{}
				return isSvc
			},
			Method: func(resolver *oauth20.Resolver, ctx context.Context, id string) (graphql.SystemAuth, error) {
				return resolver.RequestClientCredentialsForApplication(ctx, id)
			},
		},
		{
			Name:     "Integration System",
			IntSysID: &id,
			ObjType:  pkgmodel.IntegrationSystemReference,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				return rtmSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			IntegrationSystemServiceFn: func() *automock.IntegrationSystemService {
				isSvc := &automock.IntegrationSystemService{}
				isSvc.On("Exists", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return isSvc
			},
			Method: func(resolver *oauth20.Resolver, ctx context.Context, id string) (graphql.SystemAuth, error) {
				return resolver.RequestClientCredentialsForIntegrationSystem(ctx, id)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			modelSystemAuth := fixModelSystemAuth(clientID, testCase.RtmID, testCase.AppID, testCase.IntSysID)

			persist, transact := txGen.ThatSucceeds()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)

			svc := &automock.Service{}
			svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), testCase.ObjType).Return(credsData, nil).Once()
			defer svc.AssertExpectations(t)

			rtmSvc := testCase.RuntimeServiceFn()
			defer rtmSvc.AssertExpectations(t)

			appSvc := testCase.ApplicationServiceFn()
			defer appSvc.AssertExpectations(t)

			isSvc := testCase.IntegrationSystemServiceFn()
			defer isSvc.AssertExpectations(t)

			systemAuthSvc := &automock.SystemAuthService{}
			systemAuthSvc.On("CreateWithCustomID", txtest.CtxWithDBMatcher(), clientID, testCase.ObjType, id, authInput).Return(clientID, nil).Once()
			systemAuthSvc.On("GetByIDForObject", txtest.CtxWithDBMatcher(), testCase.ObjType, clientID).Return(modelSystemAuth, nil).Once()
			defer systemAuthSvc.AssertExpectations(t)

			systemAuthConv := &automock.SystemAuthConverter{}
			systemAuthConv.On("ToGraphQL", modelSystemAuth).Return(expectedResult, nil).Once()
			defer systemAuthConv.AssertExpectations(t)

			resolver := oauth20.NewResolver(transact, svc, appSvc, rtmSvc, isSvc, systemAuthSvc, systemAuthConv)

			// When
			result, err := testCase.Method(resolver, context.TODO(), id)

			// Then
			assert.Equal(t, expectedResult, result)
			assert.Nil(t, err)
		})
	}
}

func TestResolver_CommonRequestClientCredentialsError(t *testing.T) {
	// GIVEN
	id := "foo"
	objType := pkgmodel.RuntimeReference
	clientID := "clientid"
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	modelSystemAuth := fixModelSystemAuth(clientID, &id, nil, nil)
	credsData := &model.OAuthCredentialDataInput{
		ClientID:     "clientid",
		ClientSecret: "secret",
		URL:          "url",
	}
	authInput := &model.AuthInput{Credential: &model.CredentialDataInput{Oauth: credsData}}

	testCases := []struct {
		Name                string
		ExpectedError       error
		RuntimeServiceFn    func() *automock.RuntimeService
		ServiceFn           func() *automock.Service
		SystemAuthServiceFn func() *automock.SystemAuthService
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
	}{
		{
			Name:            "Error - Transaction Commit",
			ExpectedError:   testErr,
			TransactionerFn: txGen.ThatFailsOnCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), objType).Return(credsData, nil).Once()
				svc.On("DeleteClientCredentials", txtest.CtxWithDBMatcher(), clientID).Return(nil).Once()
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("CreateWithCustomID", txtest.CtxWithDBMatcher(), clientID, objType, id, authInput).Return(clientID, nil).Once()
				systemAuthSvc.On("GetByIDForObject", txtest.CtxWithDBMatcher(), objType, clientID).Return(modelSystemAuth, nil).Once()
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Get System Auth",
			ExpectedError:   testErr,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), objType).Return(credsData, nil).Once()
				svc.On("DeleteClientCredentials", txtest.CtxWithDBMatcher(), clientID).Return(nil).Once()
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("CreateWithCustomID", txtest.CtxWithDBMatcher(), clientID, objType, id, authInput).Return(clientID, nil).Once()
				systemAuthSvc.On("GetByIDForObject", txtest.CtxWithDBMatcher(), objType, clientID).Return(nil, testErr).Once()
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Create System Auth",
			ExpectedError:   testErr,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), objType).Return(credsData, nil).Once()
				svc.On("DeleteClientCredentials", txtest.CtxWithDBMatcher(), clientID).Return(nil).Once()
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("CreateWithCustomID", txtest.CtxWithDBMatcher(), clientID, objType, id, authInput).Return("", testErr).Once()
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Multiple: Create System Auth and Delete Client",
			ExpectedError:   errors.New("2 errors occurred:\n\t* test error\n\t* test error"),
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), objType).Return(credsData, nil).Once()
				svc.On("DeleteClientCredentials", txtest.CtxWithDBMatcher(), clientID).Return(testErr).Once()
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("CreateWithCustomID", txtest.CtxWithDBMatcher(), clientID, objType, id, authInput).Return("", testErr).Once()
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Create Client",
			ExpectedError:   testErr,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), objType).Return(nil, testErr).Once()
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Empty Credentials",
			ExpectedError:   errors.New("client credentials cannot be empty"),
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(true, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("CreateClientCredentials", txtest.CtxWithDBMatcher(), objType).Return(nil, nil).Once()
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Exists",
			ExpectedError:   testErr,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(false, testErr).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Doesn't Exist",
			ExpectedError:   errors.New("Runtime with ID 'foo' not found"),
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Exist", txtest.CtxWithDBMatcher(), id).Return(false, nil).Once()
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				return systemAuthSvc
			},
		},
		{
			Name:            "Error - Transaction Begin error",
			ExpectedError:   testErr,
			TransactionerFn: txGen.ThatFailsOnBegin,
			RuntimeServiceFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				return rtmSvc
			},
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			SystemAuthServiceFn: func() *automock.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				return systemAuthSvc
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)

			svc := testCase.ServiceFn()
			defer svc.AssertExpectations(t)

			rtmSvc := testCase.RuntimeServiceFn()
			defer rtmSvc.AssertExpectations(t)

			systemAuthSvc := testCase.SystemAuthServiceFn()
			defer systemAuthSvc.AssertExpectations(t)

			resolver := oauth20.NewResolver(transact, svc, nil, rtmSvc, nil, systemAuthSvc, nil)

			// When
			_, err := resolver.RequestClientCredentialsForRuntime(context.TODO(), id)

			// Then
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
		})
	}
}

func fixModelSystemAuth(clientID string, rtmID, appID, isID *string) *pkgmodel.SystemAuth {
	return &pkgmodel.SystemAuth{
		ID:                  clientID,
		TenantID:            str.Ptr(""),
		RuntimeID:           rtmID,
		IntegrationSystemID: isID,
		AppID:               appID,
		Value: &model.Auth{
			Credential: model.CredentialData{
				Oauth: &model.OAuthCredentialData{
					ClientID:     clientID,
					ClientSecret: "secret",
					URL:          "url",
				},
			},
		},
	}
}

func fixGQLSystemAuth(clientID string) graphql.SystemAuth {
	oauthCredsData := graphql.OAuthCredentialData{
		ClientID:     clientID,
		ClientSecret: "secret",
		URL:          "url",
	}
	return &graphql.IntSysSystemAuth{
		ID: "sysauth-id",
		Auth: &graphql.Auth{
			Credential: oauthCredsData,
		},
	}
}
