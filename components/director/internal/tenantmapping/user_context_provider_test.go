package tenantmapping_test

import (
	"context"
	"net/http"
	"net/textproto"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserContextProvider(t *testing.T) {
	username := "some-user"
	groupName := "some-group"
	expectedTenantID := uuid.New()
	expectedExternalTenantID := uuid.New()
	expectedScopes := []string{"application:read", "application:write"}
	userObjCtxType := "Static User"

	jwtAuthDetails := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow}

	staticGroups := tenantmapping.StaticGroups{
		{
			GroupName: groupName,
			Scopes:    expectedScopes,
		},
	}

	t.Run("returns tenant that is defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.GroupsKey:         []interface{}{groupName},
				},
			},
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		provider := tenantmapping.NewUserContextProvider(staticGroupRepoMock, tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, tenantRepoMock)
	})

	t.Run("returns tenant that is defined in the Header map of ReqData", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey): []string{expectedExternalTenantID.String()},
				},
				Extra: map[string]interface{}{
					oathkeeper.GroupsKey: []interface{}{groupName},
				},
			},
		}
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		provider := tenantmapping.NewUserContextProvider(staticGroupRepoMock, tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, tenantRepoMock)
	})

	t.Run("returns scopes defined on the StaticGroup from the request", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.GroupsKey:         []interface{}{groupName},
				},
			},
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		provider := tenantmapping.NewUserContextProvider(staticGroupRepoMock, tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
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

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.GroupsKey:         []interface{}{groupName1, groupName2},
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
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName1, groupName2}).Return(staticGroups, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		provider := tenantmapping.NewUserContextProvider(staticGroupRepoMock, tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(allExpectedGroupScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, tenantRepoMock)
	})

	t.Run("returns error when tenant is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: []byte{1, 2, 3},
					oathkeeper.GroupsKey:         []interface{}{groupName},
				},
			},
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		provider := tenantmapping.NewUserContextProvider(staticGroupRepoMock, nil)

		_, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.EqualError(t, err, "could not parse external ID for user: some-user: while parsing the value for key=tenant: Internal Server Error: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock)
	})
}

func TestUserContextProviderMatch(t *testing.T) {
	t.Run("returns ID string and JWTAuthFlow when a name is specified in the Extra map of request body", func(t *testing.T) {
		username := "some-username"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"name": username,
				},
			},
		}

		provider := tenantmapping.NewUserContextProvider(nil, nil)

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
	})

	t.Run("returns error when username is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: []byte{1, 2, 3},
				},
			},
		}

		provider := tenantmapping.NewUserContextProvider(nil, nil)

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.EqualError(t, err, "while parsing the value for name: Internal Server Error: unable to cast the value to a string type")
	})

	t.Run("returns ID string and JWTAuthFlow when username attribute is specified in the Extra map of request body and no authenticators match", func(t *testing.T) {
		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		authenticatorName := "auth1"
		username := "some-username"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					authenticator.CoordinatesKey: authenticator.Coordinates{
						Name:  authenticatorName,
						Index: 0,
					},
					uniqueAttributeKey:   uniqueAttributeValue,
					identityAttributeKey: username,
				},
			},
		}

		reqData.Body.Extra[authenticator.CoordinatesKey] = authenticator.Coordinates{
			Name: "unknown",
		}
		reqData.Body.Extra[oathkeeper.UsernameKey] = username

		provider := tenantmapping.NewUserContextProvider(nil, nil)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
	})

	t.Run("return nil when does not match", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
		}

		provider := tenantmapping.NewUserContextProvider(nil, nil)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.NoError(t, err)
		require.Nil(t, authDetails)
	})
}

func getStaticGroupRepoMock() *automock.StaticGroupRepository {
	repo := &automock.StaticGroupRepository{}
	return repo
}
