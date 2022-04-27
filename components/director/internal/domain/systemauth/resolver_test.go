package systemauth_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
)

var contextParam = txtest.CtxWithDBMatcher()

func TestResolver_GenericDeleteSystemAuth(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	id := "foo"
	objectID := "bar"
	objectType := pkgmodel.RuntimeReference
	modelSystemAuth := fixModelSystemAuth(id, objectType, objectID, fixModelAuth())
	oauthModelSystemAuth := fixModelSystemAuth(id, objectType, objectID, &model.Auth{
		Credential: model.CredentialData{
			Oauth: &model.OAuthCredentialData{
				ClientID:     "clientid",
				ClientSecret: "clientsecret",
				URL:          "foo.bar/token",
			},
		},
	})
	gqlSystemAuth := fixGQLIntSysSystemAuth(id, fixGQLAuth(), objectID)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.SystemAuthService
		OAuthServiceFn     func() *automock.OAuth20Service
		OneTimeTokenFn     func() *automock.OneTimeTokenService
		ConverterFn        func() *automock.SystemAuthConverter
		AuthConverterFn    func() *automock.AuthConverter
		ExpectedSystemAuth graphql.SystemAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success - Basic Auth",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByIDForObject", contextParam, objectType, id).Return(modelSystemAuth, nil).Once()
				svc.On("DeleteByIDForObject", contextParam, objectType, id).Return(nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Success - OAuth",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByIDForObject", contextParam, objectType, id).Return(oauthModelSystemAuth, nil).Once()
				svc.On("DeleteByIDForObject", contextParam, objectType, id).Return(nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteClientCredentials", contextParam, oauthModelSystemAuth.Value.Credential.Oauth.ClientID).Return(nil)
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", oauthModelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error - GetByIDForObject SystemAuth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByIDForObject", contextParam, objectType, id).Return(nil, testErr).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Delete from DB",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByIDForObject", contextParam, objectType, id).Return(modelSystemAuth, nil).Once()
				svc.On("DeleteByIDForObject", contextParam, objectType, id).Return(testErr).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Delete Client",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByIDForObject", contextParam, objectType, id).Return(oauthModelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteClientCredentials", contextParam, oauthModelSystemAuth.Value.Credential.Oauth.ClientID).Return(testErr)
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", oauthModelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedSystemAuth: nil,
			ExpectedErr:        errors.Wrap(testErr, "while deleting OAuth 2.0 client"),
		},
		{
			Name:            "Error - Transaction Begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Transaction Commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByIDForObject", contextParam, objectType, id).Return(oauthModelSystemAuth, nil).Once()
				svc.On("DeleteByIDForObject", contextParam, objectType, id).Return(nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteClientCredentials", contextParam, oauthModelSystemAuth.Value.Credential.Oauth.ClientID).Return(nil)
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", oauthModelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)
			svc := testCase.ServiceFn()
			defer svc.AssertExpectations(t)
			oauthSvc := testCase.OAuthServiceFn()
			defer oauthSvc.AssertExpectations(t)
			converter := testCase.ConverterFn()
			defer converter.AssertExpectations(t)
			authConverter := testCase.AuthConverterFn()
			defer authConverter.AssertExpectations(t)
			ottSvc := testCase.OneTimeTokenFn()
			defer ottSvc.AssertExpectations(t)

			resolver := systemauth.NewResolver(transact, svc, oauthSvc, ottSvc, converter, authConverter)

			// WHEN
			fn := resolver.GenericDeleteSystemAuth(objectType)
			result, err := fn(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedSystemAuth, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}

func TestResolver_SystemAuth(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	id := "foo"
	objectID := "bar"
	objectType := pkgmodel.RuntimeReference
	modelSystemAuth := fixModelSystemAuth(id, objectType, objectID, fixModelAuth())
	gqlSystemAuth := fixGQLIntSysSystemAuth(id, fixGQLAuth(), objectID)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.SystemAuthService
		OAuthServiceFn     func() *automock.OAuth20Service
		OneTimeTokenFn     func() *automock.OneTimeTokenService
		ConverterFn        func() *automock.SystemAuthConverter
		AuthConverterFn    func() *automock.AuthConverter
		InputID            string
		ExpectedSystemAuth graphql.SystemAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			InputID:            id,
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error - Transaction Begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "GetGlobal")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			InputID:            id,
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Transaction Commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - GetGlobal",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - ToGraphQL",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: unusedOneTimeTokenSvc,
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(nil, testErr).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)
			svc := testCase.ServiceFn()
			defer svc.AssertExpectations(t)
			oauthSvc := testCase.OAuthServiceFn()
			defer oauthSvc.AssertExpectations(t)
			converter := testCase.ConverterFn()
			defer converter.AssertExpectations(t)
			authConverter := testCase.AuthConverterFn()
			defer authConverter.AssertExpectations(t)
			ottSvc := testCase.OneTimeTokenFn()
			defer ottSvc.AssertExpectations(t)

			resolver := systemauth.NewResolver(transact, svc, oauthSvc, ottSvc, converter, authConverter)

			// WHEN
			result, err := resolver.SystemAuth(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedSystemAuth, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}

func TestResolver_SystemAuthByToken(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	id := "foo"
	token := "token"
	objectID := "bar"
	objectType := pkgmodel.RuntimeReference
	modelSystemAuth := fixModelSystemAuth(id, objectType, objectID, fixModelAuth())
	gqlSystemAuth := fixGQLIntSysSystemAuth(id, fixGQLAuth(), objectID)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.SystemAuthService
		OAuthServiceFn     func() *automock.OAuth20Service
		OneTimeTokenFn     func() *automock.OneTimeTokenService
		ConverterFn        func() *automock.SystemAuthConverter
		AuthConverterFn    func() *automock.AuthConverter
		Token              string
		ExpectedSystemAuth graphql.SystemAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByToken", contextParam, token).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(true, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			Token:              token,
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error - Transaction Begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "GetByToken")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.AssertNotCalled(t, "IsTokenValid")
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			Token:              token,
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Transaction Commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByToken", contextParam, token).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.AssertNotCalled(t, "IsTokenValid")
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			Token:       token,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - GetByToken",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByToken", contextParam, token).Return(nil, testErr).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.AssertNotCalled(t, "IsTokenValid")
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			Token:       token,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - IsTokenValid",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByToken", contextParam, token).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(false, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				//conv.On("ToGraphQL", modelSystemAuth).Return(nil, testErr).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			Token:       token,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - ToGraphQL",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetByToken", contextParam, token).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(true, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(nil, testErr).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			Token:       token,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)
			svc := testCase.ServiceFn()
			defer svc.AssertExpectations(t)
			oauthSvc := testCase.OAuthServiceFn()
			defer oauthSvc.AssertExpectations(t)
			converter := testCase.ConverterFn()
			defer converter.AssertExpectations(t)
			authConverter := testCase.AuthConverterFn()
			defer authConverter.AssertExpectations(t)
			ottSvc := testCase.OneTimeTokenFn()
			defer ottSvc.AssertExpectations(t)

			resolver := systemauth.NewResolver(transact, svc, oauthSvc, ottSvc, converter, authConverter)

			// WHEN
			result, err := resolver.SystemAuthByToken(context.TODO(), testCase.Token)

			// then
			assert.Equal(t, testCase.ExpectedSystemAuth, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}

func TestResolver_UpdateSystemAuth(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	id := "foo"
	objectID := "bar"
	objectType := pkgmodel.RuntimeReference
	modelSystemAuth := fixModelSystemAuth(id, objectType, objectID, fixModelAuth())
	gqlSystemAuth := fixGQLIntSysSystemAuth(id, fixGQLAuth(), objectID)
	inputAuth := fixGQLAuthInput()
	modelAuth := fixModelAuth()

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.SystemAuthService
		OAuthServiceFn     func() *automock.OAuth20Service
		OneTimeTokenFn     func() *automock.OneTimeTokenService
		ConverterFn        func() *automock.SystemAuthConverter
		AuthConverterFn    func() *automock.AuthConverter
		ID                 string
		InputAuth          graphql.AuthInput
		ExpectedSystemAuth graphql.SystemAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("UpdateValue", contextParam, id, modelAuth).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ModelFromGraphQLInput", *inputAuth).Return(modelAuth, nil).Once()
				return conv
			},
			ID:                 id,
			InputAuth:          *inputAuth,
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error - Transaction Begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "UpdateValue")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.AssertNotCalled(t, "ModelFromGraphQLInput")
				return conv
			},
			ID:                 id,
			InputAuth:          *inputAuth,
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Transaction Commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("UpdateValue", contextParam, id, modelAuth).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ModelFromGraphQLInput", *inputAuth).Return(modelAuth, nil).Once()
				return conv
			},
			ID:          id,
			InputAuth:   *inputAuth,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - ModelFromGraphQLInput",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "UpdateValue")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ModelFromGraphQLInput", *inputAuth).Return(nil, testErr).Once()
				return conv
			},
			ID:          id,
			InputAuth:   *inputAuth,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - UpdateValue",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("UpdateValue", contextParam, id, modelAuth).Return(nil, testErr).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ModelFromGraphQLInput", *inputAuth).Return(modelAuth, nil).Once()
				return conv
			},
			ID:          id,
			InputAuth:   *inputAuth,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - ToGraphQL",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("UpdateValue", contextParam, id, modelAuth).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(nil, testErr).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ModelFromGraphQLInput", *inputAuth).Return(modelAuth, nil).Once()
				return conv
			},
			ID:          id,
			InputAuth:   *inputAuth,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)
			svc := testCase.ServiceFn()
			defer svc.AssertExpectations(t)
			oauthSvc := testCase.OAuthServiceFn()
			defer oauthSvc.AssertExpectations(t)
			converter := testCase.ConverterFn()
			defer converter.AssertExpectations(t)
			authConverter := testCase.AuthConverterFn()
			defer authConverter.AssertExpectations(t)
			ottSvc := testCase.OneTimeTokenFn()
			defer ottSvc.AssertExpectations(t)

			resolver := systemauth.NewResolver(transact, svc, oauthSvc, ottSvc, converter, authConverter)

			// WHEN
			result, err := resolver.UpdateSystemAuth(context.TODO(), testCase.ID, testCase.InputAuth)

			// then
			assert.Equal(t, testCase.ExpectedSystemAuth, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}

func TestResolver_InvalidateSystemAuthOneTimeToken(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	id := "foo"
	objectID := "bar"
	objectType := pkgmodel.RuntimeReference
	modelSystemAuth := fixModelSystemAuth(id, objectType, objectID, fixModelAuth())
	gqlSystemAuth := fixGQLIntSysSystemAuth(id, fixGQLAuth(), objectID)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.SystemAuthService
		OAuthServiceFn     func() *automock.OAuth20Service
		OneTimeTokenFn     func() *automock.OneTimeTokenService
		ConverterFn        func() *automock.SystemAuthConverter
		AuthConverterFn    func() *automock.AuthConverter
		ID                 string
		InputAuth          graphql.AuthInput
		ExpectedSystemAuth graphql.SystemAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.On("InvalidateToken", contextParam, id).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(true, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:                 id,
			ExpectedSystemAuth: gqlSystemAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error - Transaction Begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "GetGlobal")
				svc.AssertNotCalled(t, "InvalidateToken")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.AssertNotCalled(t, "IsTokenValid")
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:                 id,
			ExpectedSystemAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error - Transaction Commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.On("InvalidateToken", contextParam, id).Return(modelSystemAuth, nil).Once()
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(true, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:          id,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - GetGlobal",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(nil, testErr).Once()
				svc.AssertNotCalled(t, "InvalidateToken")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.AssertNotCalled(t, "IsTokenValid")
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:          id,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - IsTokenValid",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.AssertNotCalled(t, "InvalidateToken")
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(false, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:          id,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - IsTokenValid",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.On("InvalidateToken", contextParam, id).Return(nil, testErr)
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(true, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:          id,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error - ToGraphQL",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("GetGlobal", contextParam, id).Return(modelSystemAuth, nil).Once()
				svc.On("InvalidateToken", contextParam, id).Return(modelSystemAuth, nil)
				return svc
			},
			OAuthServiceFn: func() *automock.OAuth20Service {
				return &automock.OAuth20Service{}
			},
			OneTimeTokenFn: func() *automock.OneTimeTokenService {
				svc := &automock.OneTimeTokenService{}
				svc.On("IsTokenValid", modelSystemAuth).Return(true, nil)
				return svc
			},
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(nil, testErr).Once()
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ID:          id,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			defer persist.AssertExpectations(t)
			defer transact.AssertExpectations(t)
			svc := testCase.ServiceFn()
			defer svc.AssertExpectations(t)
			oauthSvc := testCase.OAuthServiceFn()
			defer oauthSvc.AssertExpectations(t)
			converter := testCase.ConverterFn()
			defer converter.AssertExpectations(t)
			authConverter := testCase.AuthConverterFn()
			defer authConverter.AssertExpectations(t)
			ottSvc := testCase.OneTimeTokenFn()
			defer ottSvc.AssertExpectations(t)

			resolver := systemauth.NewResolver(transact, svc, oauthSvc, ottSvc, converter, authConverter)

			// WHEN
			result, err := resolver.InvalidateSystemAuthOneTimeToken(context.TODO(), testCase.ID)

			// then
			assert.Equal(t, testCase.ExpectedSystemAuth, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			}
		})
	}
}

func unusedOneTimeTokenSvc() *automock.OneTimeTokenService {
	return &automock.OneTimeTokenService{}
}
