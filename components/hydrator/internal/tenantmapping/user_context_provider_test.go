package tenantmapping_test

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserContextProvider(t *testing.T) {
	username := "some-user"
	expectedTenantID := uuid.New()
	expectedExternalTenantID := uuid.New()
	expectedScopes := []string{"application:read", "application:write"}
	userObjCtxType := "Static User"

	jwtAuthDetails := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow}

	t.Run("returns tenant and scopes that are defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.ScopesKey:         strings.Join(expectedScopes, " "),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, directorClientMock)
	})

	t.Run("returns tenant and scopes that are defined in the Header map of ReqData", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey): []string{expectedExternalTenantID.String()},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ScopesKey):         []string{strings.Join(expectedScopes, " ")},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, directorClientMock)
	})

	t.Run("returns tenant which is defined in the Extra map and scopes which is defined in the Header map of ReqData", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ScopesKey): []string{strings.Join(expectedScopes, " ")},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, directorClientMock)
	})

	t.Run("returns tenant which is defined in the Header map and scopes which is defined in the Extra map of ReqData", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: strings.Join(expectedScopes, " "),
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey): []string{expectedExternalTenantID.String()},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, directorClientMock)
	})

	t.Run("returns scopes defined on the StaticUser and tenant from the request", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, directorClientMock)
	})

	t.Run("returns scopes defined on the StaticGroup from the request", func(t *testing.T) {
		groupName := "test"
		expectedGroupScopes := []string{"tennants:read", "application:read"}

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.GroupsKey:         []interface{}{groupName},
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

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, staticGroupRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedGroupScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, directorClientMock)
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

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName1, groupName2}).Return(staticGroups, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, nil, staticGroupRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(allExpectedGroupScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, directorClientMock)
	})

	t.Run("returns scopes defined on the StaticUser when group not present from the request", func(t *testing.T) {
		groupName := "test"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.GroupsKey:         []interface{}{groupName},
				},
			},
		}
		var staticGroups tenantmapping.StaticGroups

		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, staticGroupRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticGroupRepoMock, directorClientMock)
	})

	t.Run("returns error when tenant from the request does not match any tenants assigned to the static user", func(t *testing.T) {
		nonExistingExternalTenantID := uuid.New().String()
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: nonExistingExternalTenantID,
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}
		testTenant := &graphql.Tenant{
			ID:         nonExistingExternalTenantID,
			InternalID: uuid.New().String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		directorClientMock := getDirectorClientMock()
		directorClientMock.On("GetTenantByExternalID", mock.Anything, nonExistingExternalTenantID).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		_, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.EqualError(t, err, apperrors.NewInternalError(fmt.Sprintf("Static tenant with username: some-user missmatch external tenant: %s", nonExistingExternalTenantID)).Error())

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, directorClientMock)
	})

	t.Run("returns error when tenant is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: []byte{1, 2, 3},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		directorClientMock := getDirectorClientMock()
		//directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		_, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.EqualError(t, err, "could not parse external ID for user: some-user: while parsing the value for key=tenant: Internal Server Error: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when scopes is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: []byte{1, 2, 3},
				},
			},
		}
		staticUser := tenantmapping.StaticUser{
			Username: username,
			Tenants:  []string{expectedExternalTenantID.String()},
			Scopes:   expectedScopes,
		}

		directorClientMock := getDirectorClientMock()
		//directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		_, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetails)

		require.EqualError(t, err, "while getting user data for user: some-user: while fetching scopes: while parsing the value for scope: Internal Server Error: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when user repository returns error", func(t *testing.T) {
		reqData := oathkeeper.ReqData{}
		username := "non-existing"

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(tenantmapping.StaticUser{}, errors.New("some-error")).Once()

		directorClientMock := getDirectorClientMock()
		//directorClientMock.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		provider := tenantmapping.NewUserContextProvider(directorClientMock, staticUserRepoMock, nil)

		jwtAuthDetailsWithMissingUser := jwtAuthDetails
		jwtAuthDetailsWithMissingUser.AuthID = username
		_, err := provider.GetObjectContext(context.TODO(), reqData, jwtAuthDetailsWithMissingUser)

		require.EqualError(t, err, "while getting user data for user: non-existing: while searching for a static user with username non-existing: some-error")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
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

		provider := tenantmapping.NewUserContextProvider(nil, nil, nil)

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

		provider := tenantmapping.NewUserContextProvider(nil, nil, nil)

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

		provider := tenantmapping.NewUserContextProvider(nil, nil, nil)
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

		provider := tenantmapping.NewUserContextProvider(nil, nil, nil)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.NoError(t, err)
		require.Nil(t, authDetails)
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
