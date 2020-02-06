package tenantmapping_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"

	"github.com/google/uuid"
	systemauthmock "github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	tenantmappingmock "github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"
)

func TestMapperForSystemAuthGetObjectContext(t *testing.T) {
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
		reqData := tenantmapping.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.application").Return(expectedScopes, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, scopesGetterMock, nil)

		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.CertificateFlow)

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
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID,
					tenantmapping.ScopesKey:         expectedScopes,
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, tenantRepoMock)

		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

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
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ScopesKey: expectedScopes,
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

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

		reqData := tenantmapping.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(&model.SystemAuth{}, errors.New("some-error")).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while retrieving system auth from database: some-error")

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when unable to get the ReferenceObjectType of underlying SystemAuth", func(t *testing.T) {
		authID := uuid.New()
		sysAuth := &model.SystemAuth{}

		reqData := tenantmapping.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while getting reference object type: unknown reference object type")

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when unable to get the scopes from the ReqData in the Integration System SystemAuth case", func(t *testing.T) {
		authID := uuid.New()
		refObjID := uuid.New()

		sysAuth := &model.SystemAuth{
			ID:                  authID.String(),
			IntegrationSystemID: str.Ptr(refObjID.String()),
		}
		reqData := tenantmapping.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while fetching the tenant and scopes for object of type Integration System: while fetching scopes: the key (scope) does not exist in source object")

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

		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: []byte{1, 2, 3},
					tenantmapping.ScopesKey:         "application:read",
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while fetching the tenant and scopes for object of type Application: while fetching tenant: while parsing the value for tenant: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, systemAuthSvcMock)
	})

	t.Run("returns error when underlying tenant specified in the ReqData differs from the on defined in SystemAuth in the Application or Runtime SystemAuth case", func(t *testing.T) {
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
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: externalTenantID,
					tenantmapping.ScopesKey:         "application:read",
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, externalTenantID).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, tenantRepoMock)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while fetching the tenant and scopes for object of type Application: tenant mismatch")

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
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: tenant2ID.String(),
					tenantmapping.ScopesKey:         "application:read",
				},
			},
		}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while fetching the tenant and scopes for object of type Application: system auth tenant id cannot be nil")

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
		reqData := tenantmapping.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, nil, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.OAuth2Flow)

		require.EqualError(t, err, "while fetching the tenant and scopes for object of type Application: while fetching scopes: the key (scope) does not exist in source object")

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
		reqData := tenantmapping.ReqData{}

		systemAuthSvcMock := getSystemAuthSvcMock()
		systemAuthSvcMock.On("GetGlobal", mock.Anything, authID.String()).Return(sysAuth, nil).Once()

		scopesGetterMock := getScopesGetterMock()
		scopesGetterMock.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.application").Return([]string{}, errors.New("some-error")).Once()

		mapper := tenantmapping.NewMapperForSystemAuth(systemAuthSvcMock, scopesGetterMock, nil)

		_, err := mapper.GetObjectContext(context.TODO(), reqData, authID.String(), tenantmapping.CertificateFlow)

		require.EqualError(t, err, "while fetching the tenant and scopes for object of type Application: while fetching scopes: some-error")

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
