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

func TestMapperForUserGetTenantAndScopes(t *testing.T) {
	t.Run("GetTenantAndScopes returns tenant and scopes that are defined in the Extra map of ReqData", func(t *testing.T) {
		expectedTenantID := uuid.New()
		expectedScopes := "application:read"
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.TenantKey: expectedTenantID.String(),
				tenantmapping.ScopesKey: expectedScopes,
			},
		}
		mapper := tenantmapping.NewMapperForUser(nil)
		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, "admin")

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, expectedScopes, scopes)
	})

	t.Run("GetTenantAndScopes returns tenant and scopes that are defined in the Header map of ReqData", func(t *testing.T) {
		expectedTenantID := uuid.New()
		expectedScopes := "application:read"
		reqData := tenantmapping.ReqData{
			Header: http.Header{
				textproto.CanonicalMIMEHeaderKey(tenantmapping.TenantKey): []string{expectedTenantID.String()},
				textproto.CanonicalMIMEHeaderKey(tenantmapping.ScopesKey): []string{expectedScopes},
			},
		}
		mapper := tenantmapping.NewMapperForUser(nil)
		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, "admin")

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, expectedScopes, scopes)
	})

	t.Run("GetTenantAndScopes returns tenant which is defined in the Extra map and scopes which is defined in the Header map of ReqData", func(t *testing.T) {
		expectedTenantID := uuid.New()
		expectedScopes := "application:read"
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.TenantKey: expectedTenantID.String(),
			},
			Header: http.Header{
				textproto.CanonicalMIMEHeaderKey(tenantmapping.ScopesKey): []string{expectedScopes},
			},
		}
		mapper := tenantmapping.NewMapperForUser(nil)
		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, "admin")

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, expectedScopes, scopes)
	})

	t.Run("GetTenantAndScopes returns tenant which is defined in the Header map and scopes which is defined in the Extra map of ReqData", func(t *testing.T) {
		expectedTenantID := uuid.New()
		expectedScopes := "application:read"
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.ScopesKey: expectedScopes,
			},
			Header: http.Header{
				textproto.CanonicalMIMEHeaderKey(tenantmapping.TenantKey): []string{expectedTenantID.String()},
			},
		}
		mapper := tenantmapping.NewMapperForUser(nil)
		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, "admin")

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, expectedScopes, scopes)
	})

	t.Run("GetTenantAndScopes returns error when tenant is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.TenantKey: []byte{1, 2, 3},
			},
		}
		mapper := tenantmapping.NewMapperForUser(nil)
		_, _, err := mapper.GetTenantAndScopes(reqData, "admin")

		require.EqualError(t, err, "while fetching tenant: while parsing the value for tenant: unable to cast the value to a string type")
	})

	t.Run("GetTenantAndScopes returns error when scopes is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.ScopesKey: []byte{1, 2, 3},
			},
		}
		mapper := tenantmapping.NewMapperForUser(nil)
		_, _, err := mapper.GetTenantAndScopes(reqData, "admin")

		require.EqualError(t, err, "while fetching scopes: while parsing the value for scope: unable to cast the value to a string type")
	})

	t.Run("GetTenantAndScopes returns error when no tenant and scopes are defined in the request and user repo returns error", func(t *testing.T) {
		reqData := tenantmapping.ReqData{}
		username := "non-existing"

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(tenantmapping.StaticUser{}, errors.New("some-error")).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock)

		_, _, err := mapper.GetTenantAndScopes(reqData, username)

		require.EqualError(t, err, "while searching for a static user with username non-existing: some-error")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("GetTenantAndScopes returns tenant and scopes defined on the StaticUser when both are not defined in the request", func(t *testing.T) {
		reqData := tenantmapping.ReqData{}
		username := "some-user"
		expectedTenantID := uuid.New()
		expectedScopes := []string{"application:read"}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenant:   expectedTenantID,
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock)

		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, strings.Join(expectedScopes, " "), scopes)

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("GetTenantAndScopes returns tenant defined on the StaticUser and scopes from the request", func(t *testing.T) {
		username := "some-user"
		expectedTenantID := uuid.New()
		expectedScopes := []string{"runtime:read"}
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.ScopesKey: strings.Join(expectedScopes, " "),
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenant:   expectedTenantID,
			Scopes:   []string{"application:read"},
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock)

		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, strings.Join(expectedScopes, " "), scopes)

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("GetTenantAndScopes returns scopes defined on the StaticUser and tenant from the request", func(t *testing.T) {
		username := "some-user"
		expectedTenantID := uuid.New()
		expectedScopes := []string{"application:read"}
		reqData := tenantmapping.ReqData{
			Extra: map[string]interface{}{
				tenantmapping.TenantKey: expectedTenantID.String(),
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenant:   uuid.New(),
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock)

		tenant, scopes, err := mapper.GetTenantAndScopes(reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), tenant)
		require.Equal(t, strings.Join(expectedScopes, " "), scopes)

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})
}

func getStaticUserRepoMock() *automock.StaticUserRepository {
	repo := &automock.StaticUserRepository{}
	return repo
}
