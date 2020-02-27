package tenantmapping_test

import (
	"context"
	"net/http"
	"net/textproto"

	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMapperForUserGetObjectContext(t *testing.T) {
	username := "some-user"
	expectedTenantID := uuid.New()
	expectedExternalTenantID := uuid.New()
	expectedScopes := []string{"application:read", "application:write"}
	userObjCtxType := "Static User"

	t.Run("returns tenant and scopes that are defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID.String(),
					tenantmapping.ScopesKey:         strings.Join(expectedScopes, " "),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns tenant and scopes that are defined in the Header map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ExternalTenantKey): []string{expectedExternalTenantID.String()},
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ScopesKey):         []string{strings.Join(expectedScopes, " ")},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns tenant which is defined in the Extra map and scopes which is defined in the Header map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID.String(),
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ScopesKey): []string{strings.Join(expectedScopes, " ")},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns tenant which is defined in the Header map and scopes which is defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ScopesKey: strings.Join(expectedScopes, " "),
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(tenantmapping.ExternalTenantKey): []string{expectedExternalTenantID.String()},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns scopes defined on the StaticUser and tenant from the request", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID.String(),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns scopes defined on the StaticGroup from the request", func(t *testing.T) {
		groupName := "test"
		expectedGroupScopes := []string{"tennants:read", "application:read"}

		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID.String(),
					tenantmapping.GroupsKey:         []interface{}{groupName},
				},
			},
		}
		staticGroups := tenantmapping.StaticGroups{
			{
				GroupName: groupName,
				Scopes:    expectedGroupScopes,
			},
		}

		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", []string{groupName}).Return(staticGroups, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, staticGroupRepoMock, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedGroupScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, tenantRepoMock)
	})

	t.Run("returns all unique scopes defined on the StaticGroups from the request", func(t *testing.T) {
		groupName1 := "test"
		groupName2 := "test2"
		expectedGroupScopes := []string{"tennants:read", "application:read"}
		expectedGroupScopes2 := []string{"application:read", "applications:edit"}
		allExpectedGroupScopes := []string{"tennants:read", "application:read", "applications:edit"}

		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID.String(),
					tenantmapping.GroupsKey:         []interface{}{groupName1, groupName2},
				},
			},
		}
		staticGroups := tenantmapping.StaticGroups{
			{
				GroupName: groupName1,
				Scopes:    expectedGroupScopes,
			}, {
				GroupName: groupName2,
				Scopes:    expectedGroupScopes2,
			},
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", []string{groupName1, groupName2}).Return(staticGroups, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticGroupRepoMock, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(allExpectedGroupScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, tenantRepoMock)
	})

	t.Run("returns scopes defined on the StaticUser when group not present from the request", func(t *testing.T) {
		groupName := "test"
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: expectedExternalTenantID.String(),
					tenantmapping.GroupsKey:         []interface{}{groupName},
				},
			},
		}
		var staticGroups tenantmapping.StaticGroups

		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", []string{groupName}).Return(staticGroups, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, staticGroupRepoMock, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, tenantRepoMock)
	})

	t.Run("returns error when tenant from the request does not match any tenants assigned to the static user", func(t *testing.T) {
		nonExistingExternalTenantID := uuid.New().String()
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: nonExistingExternalTenantID,
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             uuid.New().String(),
			ExternalTenant: nonExistingExternalTenantID,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, nonExistingExternalTenantID).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, tenantRepoMock)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "tenant mismatch")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns error when tenant is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := tenantmapping.ReqData{
			Body: tenantmapping.ReqBody{
				Extra: map[string]interface{}{
					tenantmapping.ExternalTenantKey: []byte{1, 2, 3},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, nil)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "while fetching external tenant: while parsing the value for tenant: unable to cast the value to a string type")

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
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, nil)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "while getting user data: while fetching scopes: while parsing the value for scope: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when user repository returns error", func(t *testing.T) {
		reqData := tenantmapping.ReqData{}
		username := "non-existing"

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(tenantmapping.StaticUser{}, errors.New("some-error")).Once()

		mapper := tenantmapping.NewMapperForUser(staticUserRepoMock, nil, nil)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "while getting user data: while searching for a static user with username non-existing: some-error")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})
}

func getStaticUserRepoMock() *automock.StaticUserRepository {
	repo := &automock.StaticUserRepository{}
	return repo
}

func getStaticGroupRepoMock() *automock.StaticGroupRepository {
	repo := &automock.StaticGroupRepository{}
	return repo
}
