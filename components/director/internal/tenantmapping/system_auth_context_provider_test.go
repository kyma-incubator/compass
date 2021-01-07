package tenantmapping_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"

	"github.com/google/uuid"
	systemauthmock "github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	tenantmappingmock "github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"
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

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.application").Return(expectedScopes, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, scopesGetterMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, "", objCtx.ExternalTenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock, scopesGetterMock)
	})

	t.Run("returns tenant and scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		expectedTenantID := uuid.New()
		expectedExternalTenantID := uuid.New().String()
		expectedScopes := "application:read"
		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID,
		}
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID,
					oathkeeper.ScopesKey:         expectedScopes,
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID).Return(tenantMappingModel, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, tenantRepoMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, expectedExternalTenantID, objCtx.ExternalTenantID)
		require.Equal(t, expectedScopes, objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Integration System", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock, tenantRepoMock)
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

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, "", objCtx.ExternalTenantID)
		require.Equal(t, expectedScopes, objCtx.Scopes)
		require.Equal(t, refObjID.String(), objCtx.ConsumerID)
		require.Equal(t, "Application", string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when unable to get SystemAuth from the service", func(t *testing.T) {
		authID := uuid.New()

		reqData := oathkeeper.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(&model.SystemAuth{}, errors.New("some-error")).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, "while retrieving system auth from database: some-error")

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when unable to get the ReferenceObjectType of underlying SystemAuth", func(t *testing.T) {
		authID := uuid.New()
		sysAuth := &model.SystemAuth{}
		sysAuth.ID = "42"

		reqData := oathkeeper.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, "while getting reference object type for system auth id 42: Internal Server Error: unknown reference object type")

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when unable to get the scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}
		reqData := oathkeeper.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Integration System, using auth flow: OAuth2: while fetching scopes: the key does not exist in the source object [key=scope]", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when unable to parse tenant specified in the ReqData in the Application or Runtime SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			AppID:    str.Ptr(refObjID.String()),
			TenantID: str.Ptr("123"),
		}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: []byte{1, 2, 3},
					oathkeeper.ScopesKey:         "application:read",
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: OAuth2: while fetching tenant external id: while parsing the value for key=tenant: Internal Server Error: unable to cast the value to a string type", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns empty tenant when underlying tenant specified in the ReqData differs from the on defined in SystemAuth in the Application or Runtime SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()
		externalTenantID := uuid.New().String()
		tenant1ID := uuid.New()
		tenant2ID := uuid.New()
		sysAuth := &model.SystemAuth{
			ID:       authID.String(),
			TenantID: str.Ptr(tenant1ID.String()),
			AppID:    str.Ptr(refObjID.String()),
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             tenant2ID.String(),
			ExternalTenant: externalTenantID,
		}
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: externalTenantID,
					oathkeeper.ScopesKey:         "application:read",
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, externalTenantID).Return(tenantMappingModel, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, tenantRepoMock)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.Equal(t, objCtx.TenantID, "")
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
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

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: OAuth2: Internal Server Error: system auth tenant id cannot be nil", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
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

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, nil, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.OAuth2Flow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: OAuth2: while fetching scopes: the key does not exist in the source object [key=scope]", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
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

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.application").Return([]string{}, errors.New("some-error")).Once()

		provider := tenantmapping.NewSystemAuthContextProvider(systemAuthSvcMock, scopesGetterMock, nil)
		authDetails := oathkeeper.AuthDetails{AuthID: authID.String(), AuthFlow: oathkeeper.CertificateFlow}

		_, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)

		require.EqualError(t, err, fmt.Sprintf("while fetching the tenant and scopes for system auth with id: %s, object type: Application, using auth flow: Certificate: while fetching scopes: some-error", sysAuth.ID))

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock, scopesGetterMock)
	})
}

func getSystemAuthSvcMock() *systemauthmock.SystemAuthService {
	svc := &systemauthmock.SystemAuthService{}
	return svc
}

func getScopesGetterMock() *tenantmappingmock.ScopesGetter {
	scopesGetter := &tenantmappingmock.ScopesGetter{}
	return scopesGetter
}

func getTenantRepositoryMock() *tenantmappingmock.TenantRepository {
	tenantRepo := &tenantmappingmock.TenantRepository{}
	return tenantRepo
}
