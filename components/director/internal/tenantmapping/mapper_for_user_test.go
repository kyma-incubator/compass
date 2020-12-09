package tenantmapping_test

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
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
	expectedScopesInterfaceArray := []interface{}{"application:read", "application:write"}
	userObjCtxType := "Static User"

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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns tenant and scopes that are defined in the Extra map of ReqData in accordance with custom authenticator", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					oathkeeper.ScopesKey: expectedScopesInterfaceArray,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := []authenticator.Config{
			{
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					TenantAttribute: authenticator.Attribute{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		mapper := tenantmapping.NewMapperForUser(authn, nil, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns tenant and scopes that are defined in the Extra map of ReqData not in accordance with custom authenticator", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: expectedExternalTenantID.String(),
					oathkeeper.ScopesKey:         strings.Join(expectedScopes, " "),
					tenantAttributeKey:           expectedExternalTenantID.String(),
					"extra": map[string]interface{}{
						"unique": "something-different",
					},
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

		authn := []authenticator.Config{
			{
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					TenantAttribute: authenticator.Attribute{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		mapper := tenantmapping.NewMapperForUser(authn, staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
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
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
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
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
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
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, tenantRepoMock)
		objCtx, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
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
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, tenantRepoMock)
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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, staticGroupRepoMock, tenantRepoMock)
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

		mapper := tenantmapping.NewMapperForUser(nil, nil, staticGroupRepoMock, tenantRepoMock)
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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		staticGroupRepoMock := getStaticGroupRepoMock()
		staticGroupRepoMock.On("Get", mock.Anything, []string{groupName}).Return(staticGroups, nil).Once()

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, staticGroupRepoMock, tenantRepoMock)
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
		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             uuid.New().String(),
			ExternalTenant: nonExistingExternalTenantID,
		}

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, nonExistingExternalTenantID).Return(tenantMappingModel, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, tenantRepoMock)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, apperrors.NewInternalError(fmt.Sprintf("Static tenant with username: some-user missmatch external tenant: %s", nonExistingExternalTenantID)).Error())

		mock.AssertExpectationsForObjects(t, staticUserRepoMock, tenantRepoMock)
	})

	t.Run("returns error when some scopes that are defined in the Extra map of ReqData are not strings when custom authenticator is activated", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					oathkeeper.ScopesKey: []interface{}{"application:read", 1},
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := []authenticator.Config{
			{
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					TenantAttribute: authenticator.Attribute{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		mapper := tenantmapping.NewMapperForUser(authn, nil, nil, tenantRepoMock)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "while parsing the value for scope: Internal Server Error: unable to cast the value to a string type")
	})

	t.Run("returns error when some tenant that is defined in the Extra map of ReqData is empty when custom authenticator is activated", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: expectedScopesInterfaceArray,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := []authenticator.Config{
			{
				Name: "test-authenticator",
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					TenantAttribute: authenticator.Attribute{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		mapper := tenantmapping.NewMapperForUser(authn, nil, nil, tenantRepoMock)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, fmt.Sprintf("tenant attribute %q missing from %s authenticator token", authn[0].Attributes.TenantAttribute.Key, authn[0].Name))
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

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, nil)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

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

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(staticUser, nil).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, nil)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "while getting user data for user: some-user: while fetching scopes: while parsing the value for scope: Internal Server Error: unable to cast the value to a string type")

		mock.AssertExpectationsForObjects(t, staticUserRepoMock)
	})

	t.Run("returns error when user repository returns error", func(t *testing.T) {
		reqData := oathkeeper.ReqData{}
		username := "non-existing"

		staticUserRepoMock := getStaticUserRepoMock()
		staticUserRepoMock.On("Get", username).Return(tenantmapping.StaticUser{}, errors.New("some-error")).Once()

		mapper := tenantmapping.NewMapperForUser(nil, staticUserRepoMock, nil, nil)
		_, err := mapper.GetObjectContext(context.TODO(), reqData, username)

		require.EqualError(t, err, "while getting user data for user: non-existing: while searching for a static user with username non-existing: some-error")

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
