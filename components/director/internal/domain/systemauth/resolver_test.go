package systemauth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var contextParam = txtest.CtxWithDBMatcher()

func TestResolver_GenericDeleteSystemAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	id := "foo"
	objectID := "bar"
	objectType := model.RuntimeReference
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
	gqlSystemAuth := fixGQLSystemAuth(id, fixGQLAuth())

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.SystemAuthService
		OAuthServiceFn     func() *automock.OAuth20Service
		ConverterFn        func() *automock.SystemAuthConverter
		ExpectedSystemAuth *graphql.SystemAuth
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", oauthModelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				return conv
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", modelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", oauthModelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				return conv
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
			ConverterFn: func() *automock.SystemAuthConverter {
				conv := &automock.SystemAuthConverter{}
				conv.On("ToGraphQL", oauthModelSystemAuth).Return(gqlSystemAuth, nil).Once()
				return conv
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

			resolver := systemauth.NewResolver(transact, svc, oauthSvc, converter)

			// when
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
