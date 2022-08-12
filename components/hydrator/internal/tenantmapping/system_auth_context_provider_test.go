package tenantmapping_test

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSystemAuthContextProvider(t *testing.T) {
	t.Run("returns tenant and scopes in the Application or Runtime SystemAuth case for Certificate flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedScopes := []string{"application:read"}

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}

		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(expectedScopes, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, scopesGetterMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, "", objCtx.ExternalTenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock, scopesGetterMock)
	})

	t.Run("returns tenant and scopes when region is found in the Application or Runtime SystemAuth case for Certificate flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := []string{"application:read"}

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID,
			InternalID: expectedTenantID.String(),
			Labels: map[string]interface{}{
				"region": "eu-1",
			},
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID).Return(testTenant, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(expectedScopes, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, scopesGetterMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, expectedExternalTenantID, objCtx.ExternalTenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock, scopesGetterMock)
	})

	t.Run("returns tenant and scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := "application:read"

		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			TenantID:            str.Ptr(expectedTenantID.String()),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID,
			InternalID: expectedTenantID.String(),
			Labels: map[string]interface{}{
				"region": "eu-1",
			},
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
					oathkeeper.ScopesKey:         expectedScopes,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID).Return(testTenant, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, expectedExternalTenantID, objCtx.ExternalTenantID)
		require.Equal(t, expectedScopes, objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Integration System", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock, directorClientMock)
	})

	t.Run("returns tenant and scopes from the ReqData in the Application or Runtime SystemAuth case for OAuth2 flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedScopes := "application:read"

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: expectedScopes,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, "", objCtx.ExternalTenantID)
		require.Equal(t, expectedScopes, objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("updates system auth with certificate common name if it is certificate flow and not already updated", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedScopes := []string{"application:read"}

		authData := &graphql.Auth{
			OneTimeToken: &graphql.OneTimeTokenForApplication{
				TokenWithURL: graphql.TokenWithURL{
					Token: "token",
				},
			},
			CertCommonName: str.Ptr(""),
		}

		authDataUpdated := &graphql.Auth{
			OneTimeToken:   nil,
			CertCommonName: str.Ptr(authID.String()),
		}

		authDataUpdatedValue, err := auth.ToModel(authDataUpdated)
		require.NoError(t, err)

		sysAuthValue, err := auth.ToModel(authData)
		require.NoError(t, err)

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
			Value:    sysAuthValue,
		}

		sysAuthUpdate := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
			Value:    authDataUpdatedValue,
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: strings.Join(expectedScopes, " "),
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("UpdateSystemAuth", mock.Anything, sysAuthUpdate).Return(director.UpdateAuthResult{}, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(expectedScopes, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, scopesGetterMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, "", objCtx.ExternalTenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock, scopesGetterMock)
	})

	t.Run("returns error when unable to get SystemAuth from Director", func(t *testing.T) {
		authID := uuid.New()

		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(nil, errors.New("some-error")).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, "while retrieving system auth from director: some-error")

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when unable to get the ReferenceObjectType of underlying SystemAuth", func(t *testing.T) {
		authID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID: authID.String(),
		}

		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		errFormatted := fmt.Sprintf("unknown reference object type for system auth with id %s", authID)
		require.EqualError(t, err, errFormatted)

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when unable to get the scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}

		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Integration System, using auth flow: OAuth2: while fetching scopes: the key does not exist in the source object [key=scope]", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns empty tenant when unable to get tenant in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := "application:read"

		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			TenantID:            str.Ptr(expectedTenantID.String()),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
					oathkeeper.ScopesKey:         expectedScopes,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, expectedExternalTenantID)).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, "", objCtx.TenantID)
		require.Equal(t, expectedExternalTenantID, objCtx.ExternalTenantID)
		require.Equal(t, expectedScopes, objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Integration System", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock, directorClientMock)
	})

	t.Run("returns error when unable to get tenant region in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := "application:read"

		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			TenantID:            str.Ptr(expectedTenantID.String()),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID,
			InternalID: expectedTenantID.String(),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
					oathkeeper.ScopesKey:         expectedScopes,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID).Return(testTenant, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("region label not found for tenant with ID: %q", expectedExternalTenantID))

		mock.AssertExpectationsForObjects(t, directorClientMock, directorClientMock)
	})

	t.Run("returns empty tenant when unable to get tenant in the Application or Runtime SystemAuth case for Certificate flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := []string{"application:read"}

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, expectedExternalTenantID)).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(expectedScopes, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, scopesGetterMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, "", objCtx.TenantID)
		require.Equal(t, expectedExternalTenantID, objCtx.ExternalTenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, directorClientMock, scopesGetterMock)
	})

	t.Run("returns error when unable to get tenant region in the Application or Runtime SystemAuth case for Certificate flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := []string{"application:read"}

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID,
			InternalID: expectedTenantID.String(),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID).Return(testTenant, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(expectedScopes, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, scopesGetterMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("region label not found for tenant with ID: %q", expectedExternalTenantID))

		mock.AssertExpectationsForObjects(t, directorClientMock, scopesGetterMock)
	})

	t.Run("returns error when unable to parse tenant specified in the ReqData in the Application or Runtime SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr("123"),
			AppID:    str.Ptr(refObjID.String()),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: []byte{1, 2, 3},
					oathkeeper.ScopesKey:         "application:read",
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: OAuth2: while fetching tenant external id: while parsing the value for key=tenant: Internal Server Error: unable to cast the value to a string type", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns empty tenant when underlying tenant specified in the ReqData differs from the on defined in SystemAuth in the Application or Runtime SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		externalTenantID := uuid.New().String()
		tenant1ID := uuid.New()
		tenant2ID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:        authID.String(),
			TenantID:  str.Ptr(tenant1ID.String()),
			RuntimeID: str.Ptr(refObjID.String()),
		}

		testTenant := &graphql.Tenant{
			ID:         externalTenantID,
			InternalID: tenant2ID.String(),
			Labels: map[string]interface{}{
				"region": "eu-1",
			},
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: externalTenantID,
					oathkeeper.ScopesKey:         "application:read",
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, externalTenantID).Return(testTenant, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.Equal(t, objCtx.TenantID, "")
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when system auth tenant id is nil", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		tenant2ID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:    authID.String(),
			AppID: str.Ptr(refObjID.String()),
		}
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: tenant2ID.String(),
					oathkeeper.ScopesKey:         "application:read",
				},
			},
		}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: OAuth2: Internal Server Error: system auth tenant id cannot be nil", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when underlying ReqData has no scopes specified in the Application or Runtime SystemAuth case for OAuth2 flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		tenant1ID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(tenant1ID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}
		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: OAuth2: while fetching scopes: the key does not exist in the source object [key=scope]", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when scopes getter fails in the Application or Runtime SystemAuth case for Certificate flow", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}
		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "scopesPerConsumerType.application").Return([]string{}, errors.New("some-error")).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, scopesGetterMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: Certificate: while fetching scopes: some-error", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, directorClientMock, scopesGetterMock)
	})
}

func TestSystemAuthContextProviderMatch(t *testing.T) {
	t.Run("returns ID string and OAuth2Flow when a client_id is specified in the Extra map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ClientIDKey: clientID,
				},
			},
		}

		provider := tenantmapping.NewSystemAuthContextProvider(nil, nil)

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.OAuth2Flow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns nil when authenticator_coordinates is specified in the Extra map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ClientIDKey:       clientID,
					oathkeeper.ScopesKey:         "application:read",
					oathkeeper.UsernameKey:       "test",
					authenticator.CoordinatesKey: "test",
				},
			},
		}

		provider := tenantmapping.NewSystemAuthContextProvider(nil, nil)

		match, _, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.NoError(t, err)
	})

	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		provider := tenantmapping.NewSystemAuthContextProvider(nil, nil)

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey): []string{clientID},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns ID string and OneTimeTokenFlow when a client-id-from-token is specified in the Header map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		provider := tenantmapping.NewSystemAuthContextProvider(nil, nil)

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDTokenKey): []string{clientID},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.OneTimeTokenFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns nil when does not match", func(t *testing.T) {
		provider := tenantmapping.NewSystemAuthContextProvider(nil, nil)
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})

	t.Run("returns error when client_id is specified in Extra map in a non-string format", func(t *testing.T) {
		provider := tenantmapping.NewSystemAuthContextProvider(nil, nil)
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ClientIDKey: []byte{1, 2, 3},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.EqualError(t, err, "while parsing the value for client_id: Internal Server Error: unable to cast the value to a string type")
	})
}

func getDirectorClientMock() *automock.DirectorClient {
	svc := &automock.DirectorClient{}
	return svc
}

func getScopesGetterMock() *automock.ScopesGetter {
	scopesGetter := &automock.ScopesGetter{}
	return scopesGetter
}
