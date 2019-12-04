package onetimetoken_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken/automock"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var URL = `http://localhost:3001/graphql`

func TestTokenService_GetOneTimeTokenForRuntime(t *testing.T) {
	runtimeID := "98cb3b05-0f27-43ea-9249-605ac74a6cf0"
	authID := "90923fe8-91bd-4070-aa31-f2ebb07a0963"

	expectedRequest := gcli.NewRequest(fmt.Sprintf(`
		mutation { generateRuntimeToken (runtimeID:"%s")
		  {
			token
		  }
		}`, authID))

	t.Run("Success runtime token", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, "tenant")
		cli := &automock.GraphQLClient{}
		expectedToken := "token"

		expected := onetimetoken.ConnectorTokenModel{RuntimeToken: onetimetoken.ConnectorToken{Token: expectedToken}}
		cli.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).
			Run(generateFakeToken(t, expected)).Return(nil).Once()

		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.RuntimeReference, runtimeID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, mockLabelRepo(), URL)

		//WHEN
		authToken, err := svc.GenerateOneTimeToken(ctx, runtimeID, model.RuntimeReference)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedToken, authToken.Token)
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})

	t.Run("Error - generating token failed", func(t *testing.T) {
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, "tenant")
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).
			Return(testErr).Once()
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.RuntimeReference, runtimeID, (*model.AuthInput)(nil)).
			Return(authID, nil)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, mockLabelRepo(), URL)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, runtimeID, model.RuntimeReference)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})

	t.Run("Error - saving auth failed", func(t *testing.T) {
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, "tenant")
		testErr := errors.New("test error")
		cli := &automock.GraphQLClient{}
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.RuntimeReference, runtimeID, (*model.AuthInput)(nil)).
			Return("", testErr)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, mockLabelRepo(), URL)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, runtimeID, model.RuntimeReference)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})
}

func TestTokenService_GetOneTimeTokenForApp(t *testing.T) {
	appID := "98cb3b05-0f27-43ea-9249-605ac74a6cf0"
	authID := "77cabc16-9fb8-4338-b252-7b404f2e6487"

	expectedRequest := gcli.NewRequest(fmt.Sprintf(`
		mutation { generateApplicationToken (appID:"%s")
		  {
			token
		  }
		}`, authID))

	t.Run("Success application token", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, "tenant")
		cli := &automock.GraphQLClient{}
		expectedToken := "token"

		expected := onetimetoken.ConnectorTokenModel{AppToken: onetimetoken.ConnectorToken{Token: expectedToken}}
		cli.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).
			Run(generateFakeToken(t, expected)).Return(nil).Once()
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.ApplicationReference, appID, (*model.AuthInput)(nil)).
			Return(authID, nil)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, mockLabelRepo(), URL)

		//WHEN
		authToken, err := svc.GenerateOneTimeToken(ctx, appID, model.ApplicationReference)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedToken, authToken.Token)
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})

	t.Run("Error - generating token failed", func(t *testing.T) {
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, "tenant")
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).
			Return(testErr).Once()
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.ApplicationReference, appID, (*model.AuthInput)(nil)).
			Return(authID, nil)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, mockLabelRepo(), URL)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, appID, model.ApplicationReference)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})
}

func generateFakeToken(t *testing.T, generated onetimetoken.ConnectorTokenModel) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*onetimetoken.ConnectorTokenModel)
		require.True(t, ok)
		*arg = generated
	}
}

func mockLabelRepo() *automock.LabelRepository {
	repo := &automock.LabelRepository{}
	repo.On("GetByKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperrors.NewNotFoundError("id"))

	return repo
}
