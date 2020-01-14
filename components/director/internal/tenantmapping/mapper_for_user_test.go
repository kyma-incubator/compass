package tenantmapping_test

import (
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMapperForUserGetObjectContext(t *testing.T) {
	username := "some-user"
	externalTenantID := uuid.New()
	internalTenantID := uuid.New()
	expectedScopes := []string{"application:read", "application:write"}
	userObjCtxType := "Static User"

	t.Run("returns mapped internal tenant and scopes that are defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.TenantKey: externalTenantID.String(),
					tenantmapping.ScopesKey: strings.Join(expectedScopes, " "),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tntStorageSvcMock := getTenantStorageServiceMock()
		tntStorageSvcMock.On("GetInternalTenant", mock.Anything, externalTenantID.String()).Return(internalTenantID.String(), nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, tntStorageSvcMock)
		objCtx, err := mapper.GetObjectContext(nil, reqData, username)

		require.NoError(t, err)
		require.Equal(t, internalTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tntStorageSvcMock)
	})

	t.Run("returns mapped internal tenant and scopes that are defined in the Header map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.TenantKey): []string{externalTenantID.String()},
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ScopesKey): []string{strings.Join(expectedScopes, " ")},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tntStorageSvcMock := getTenantStorageServiceMock()
		tntStorageSvcMock.On("GetInternalTenant", mock.Anything, externalTenantID.String()).Return(internalTenantID.String(), nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, tntStorageSvcMock)
		objCtx, err := mapper.GetObjectContext(nil, reqData, username)

		require.NoError(t, err)
		require.Equal(t, internalTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tntStorageSvcMock)
	})

	t.Run("returns mapped internal tenant which is defined in the Extra map and scopes which is defined in the Header map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.TenantKey: externalTenantID.String(),
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ScopesKey): []string{strings.Join(expectedScopes, " ")},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tntStorageSvcMock := getTenantStorageServiceMock()
		tntStorageSvcMock.On("GetInternalTenant", mock.Anything, externalTenantID.String()).Return(internalTenantID.String(), nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, tntStorageSvcMock)
		objCtx, err := mapper.GetObjectContext(nil, reqData, username)

		require.NoError(t, err)
		require.Equal(t, internalTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tntStorageSvcMock)
	})

	t.Run("returns mapped internal tenant which is defined in the Header map and scopes which is defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ScopesKey: strings.Join(expectedScopes, " "),
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.TenantKey): []string{externalTenantID.String()},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tntStorageSvcMock := getTenantStorageServiceMock()
		tntStorageSvcMock.On("GetInternalTenant", mock.Anything, externalTenantID.String()).Return(internalTenantID.String(), nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, tntStorageSvcMock)
		objCtx, err := mapper.GetObjectContext(nil, reqData, username)

		require.NoError(t, err)
		require.Equal(t, internalTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tntStorageSvcMock)
	})

	t.Run("returns scopes defined on the StaticUser and mapped internal tenant from the request", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.TenantKey: externalTenantID.String(),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tntStorageSvcMock := getTenantStorageServiceMock()
		tntStorageSvcMock.On("GetInternalTenant", mock.Anything, externalTenantID.String()).Return(internalTenantID.String(), nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, tntStorageSvcMock)
		objCtx, err := mapper.GetObjectContext(nil, reqData, username)

		require.NoError(t, err)
		require.Equal(t, internalTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tntStorageSvcMock)
	})

	t.Run("returns error when tenant is not specified in the request", func(t *testing.T) {
		reqData := tenantmapping.ReqData{}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil)
		_, err := mapper.GetObjectContext(nil, reqData, username)

		require.EqualError(t, err, "while fetching tenant: the key (tenant) does not exist in source object")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when tenant from the request does not match any tenants assigned to the static user", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.TenantKey: uuid.New().String(),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil)
		_, err := mapper.GetObjectContext(nil, reqData, username)

		require.EqualError(t, err, "tenant mismatch")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when tenant is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.TenantKey: []byte{1, 2, 3},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil)
		_, err := mapper.GetObjectContext(nil, reqData, username)

		require.EqualError(t, err, "while fetching tenant: while parsing the value for tenant: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when scopes is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ScopesKey: []byte{1, 2, 3},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []uuid.UUID{externalTenantID},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil)
		_, err := mapper.GetObjectContext(nil, reqData, username)

		require.EqualError(t, err, "while fetching scopes: while parsing the value for scope: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when user repository returns error", func(t *testing.T) {
		reqData := tenantmapping.ReqData{}
		username := "non-existing"

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(tenantmapping.StaticUser{}, errors.New("some-error")).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil)
		_, err := mapper.GetObjectContext(nil, reqData, username)

		require.EqualError(t, err, "while searching for a static user with username non-existing: some-error")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})
}

func getStaticUserRepoMock() *automock.StaticUserRepository {
	repo := &automock.StaticUserRepository{}
	return repo
}

func getTenantStorageServiceMock() *automock.TenantStorageService {
	svc := &automock.TenantStorageService{}
	return svc
}
