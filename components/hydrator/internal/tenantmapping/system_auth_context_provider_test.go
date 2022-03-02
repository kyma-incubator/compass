package tenantmapping_test

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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

		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(expectedTenantID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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

	t.Run("returns tenant and scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := "application:read"

		sysAuth := graphql.IntSysSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(expectedTenantID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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

		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(expectedTenantID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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

		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(expectedTenantID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
			Auth: &graphql.Auth{
				OneTimeToken: &graphql.OneTimeTokenForApplication{
					TokenWithURL: graphql.TokenWithURL{
						Token: "token",
					},
				},
			},
		}

		updatedAuth := &graphql.Auth{
			Credential:                      nil,
			AccessStrategy:                  nil,
			AdditionalHeaders:               nil,
			AdditionalQueryParamsSerialized: nil,
			AdditionalQueryParams:           nil,
			AdditionalHeadersSerialized:     nil,
			RequestAuth:                     nil,
			OneTimeToken:                    nil,
			CertCommonName:                  str.Ptr(authID.String()),
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
		directorClientMock.On("UpdateSystemAuth", mock.Anything, sysAuth.ID, *updatedAuth).Return(director.UpdateAuthResult{}, nil).Once()

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
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(&graphql.AppSystemAuth{}, errors.New("some-error")).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, "while retrieving system auth from director: some-error")

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when unable to get the ReferenceObjectType of underlying SystemAuth", func(t *testing.T) {
		authID := uuid.New()
		sysAuth := &graphql.AppSystemAuth{
			Type: nil,
			ID:   "42",
		}

		reqData := oathkeeper.ReqData{}

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetSystemAuthByID", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(directorClientMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		errFormatted := fmt.Sprintf("could not determine system auth type for system auth with id %s", authID)
		require.EqualError(t, err, errFormatted)

		mock.AssertExpectationsForObjects(t, directorClientMock)
	})

	t.Run("returns error when unable to get the scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()

		sysAuth := graphql.IntSysSystemAuth{
			ID:                authID.String(),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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

	t.Run("returns error when unable to parse tenant specified in the ReqData in the Application or Runtime SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()

		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr("123"),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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

		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(tenant1ID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
		}

		testTenant := &graphql.Tenant{
			ID:         externalTenantID,
			InternalID: tenant2ID.String(),
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
		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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

		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(tenant1ID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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
		sysAuth := graphql.AppSystemAuth{
			ID:                authID.String(),
			TenantID:          str.Ptr(expectedTenantID.String()),
			ReferenceObjectID: str.Ptr(refObjID.String()),
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
