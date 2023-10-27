/*
* Copyright 2020 The Compass Authors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director/automock"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorContextProvider(t *testing.T) {
	const scopePrefix = "test-compass@b12345."

	username := "some-user"
	expectedTenantID := uuid.New()
	expectedExternalTenantID := uuid.New()
	expectedScopes := []string{"application:read", "application:write"}
	prefixedScopes := []interface{}{scopePrefix + "application:read", scopePrefix + "application:write"}
	userObjCtxType := "Static User"

	jwtAuthDetails := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow, ScopePrefixes: []string{scopePrefix}}

	t.Run("returns tenant and scopes that are defined in the Extra map of ReqData", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		clientIDAttributeKey := "clientid"
		clientID := "client_id"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					clientIDAttributeKey: clientID,
					oathkeeper.ScopesKey: prefixedScopes,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
				ClientID: authenticator.Attribute{
					Key: clientIDAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, clientID, objCtx.OauthClientID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns tenant and scopes when there are multiple tenant attributes", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		tenantAttributeKey2 := "secondTenant"
		clientIDAttributeKey := "clientid"
		clientID := "client_id"
		expectedExternalTenantID2 := uuid.New()

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					tenantAttributeKey2:  expectedExternalTenantID2.String(),
					clientIDAttributeKey: clientID,
					oathkeeper.ScopesKey: prefixedScopes,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID2.String(),
			InternalID: expectedTenantID.String(),
		}

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID2.String()).Return(testTenant, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key:      tenantAttributeKey,
						Priority: 2,
					},
					{
						Key:      tenantAttributeKey2,
						Priority: 1,
					},
				},
				ClientID: authenticator.Attribute{
					Key: clientIDAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.Tenant.InternalID)
		require.Equal(t, clientID, objCtx.OauthClientID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns tenant without scopes when scopes are missing in the Extra map of ReqData", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey: expectedExternalTenantID.String(),
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.Tenant.InternalID)
		require.Equal(t, "", objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns tenant without scopes when scopes are not strings in the Extra map of ReqData", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: 1,
					tenantAttributeKey:   expectedExternalTenantID.String(),
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.Tenant.InternalID)
		require.Equal(t, "", objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns tenant and scopes without internal tenant ID when external tenant is not found", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					oathkeeper.ScopesKey: prefixedScopes,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		missingTenantErr := apperrors.NewNotFoundError(resource.Tenant, expectedExternalTenantID.String())

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(nil, missingTenantErr).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Empty(t, objCtx.Tenant.InternalID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns error when some of the scopes that are defined in the Extra map of ReqData are not strings", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					oathkeeper.ScopesKey: []interface{}{"scope1", "scope2", 123},
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		_, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.Error(t, err)
	})

	t.Run("returns error when tenant attribute is not defined in the Extra map of ReqData", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: prefixedScopes,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		testTenant := &graphql.Tenant{
			ID:         expectedExternalTenantID.String(),
			InternalID: expectedTenantID.String(),
		}

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(testTenant, nil).Once()

		authn := &authenticator.Config{
			Name: "test-authenticator",
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		_, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.EqualError(t, err, fmt.Sprintf("missing tenant attribute from %q authenticator token", authn.Name))
	})

	t.Run("returns error when external tenant id that is defined in the Extra map of ReqData cannot be resolved", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					tenantAttributeKey:   expectedExternalTenantID.String(),
					oathkeeper.ScopesKey: prefixedScopes,
					"extra": map[string]interface{}{
						"unique": uniqueAttributeValue,
					},
				},
			},
		}

		mockErr := errors.New("some-error")

		directorClient := &automock.Client{}
		directorClient.On("GetTenantByExternalID", mock.Anything, expectedExternalTenantID.String()).Return(nil, mockErr).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefixes: []string{scopePrefix},
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantsAttribute: []authenticator.TenantAttribute{
					{
						Key: tenantAttributeKey,
					},
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(directorClient, []authenticator.Config{*authn})

		_, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.EqualError(t, err, fmt.Sprintf("while getting external tenant mapping [ExternalTenantID=%s]: %s", expectedExternalTenantID, mockErr.Error()))
	})
}

func TestAuthenticatorContextProviderMatch(t *testing.T) {
	var (
		uniqueAttributeKey   string
		uniqueAttributeValue string
		identityAttributeKey string
		username             string
		region               string
		authenticatorName    string
		scopePrefixes        []string
		domainURL            string
		reqData              oathkeeper.ReqData
		authn                []authenticator.Config
	)

	setup := func() {
		uniqueAttributeKey = "uniqueAttribute"
		uniqueAttributeValue = "uniqueAttributeValue"
		identityAttributeKey = "identity"
		authenticatorName = "auth1"
		region = "region"
		scopePrefixes = []string{"prefix"}
		domainURL = "domain.com"
		username = "some-username"
		reqData = oathkeeper.ReqData{
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

		authn = []authenticator.Config{
			{
				Name: authenticatorName,
				TrustedIssuers: []authenticator.TrustedIssuer{
					{
						DomainURL:     domainURL,
						ScopePrefixes: scopePrefixes,
						Region:        region,
					},
				},
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
			},
		}
	}

	t.Run("returns ID string and JWTAuthFlow when authenticator identity is specified in the Extra map of request body", func(t *testing.T) {
		setup()
		provider := tenantmapping.NewAuthenticatorContextProvider(nil, authn)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
		require.Equal(t, scopePrefixes, authDetails.ScopePrefixes)
		require.Equal(t, region, authDetails.Region)
	})

	t.Run("returns ID string and JWTAuthFlow when multiple authenticators configured", func(t *testing.T) {
		setup()

		authn = []authenticator.Config{
			{
				Name: "emptyAuthenticator",
			},
			{
				Name: authenticatorName,
				TrustedIssuers: []authenticator.TrustedIssuer{
					{
						DomainURL:     domainURL,
						ScopePrefixes: scopePrefixes,
					},
				},
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
			},
		}
		provider := tenantmapping.NewAuthenticatorContextProvider(nil, authn)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
		require.Equal(t, scopePrefixes, authDetails.ScopePrefixes)
	})

	t.Run("returns ID string and JWTAuthFlow when authenticator identity and also default username attribute is specified in the Extra map of request body", func(t *testing.T) {
		setup()
		identityUsername := "some-identity"
		reqData.Body.Extra[oathkeeper.UsernameKey] = username
		reqData.Body.Extra[identityAttributeKey] = identityUsername

		provider := tenantmapping.NewAuthenticatorContextProvider(nil, authn)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, identityUsername, authDetails.AuthID)
		require.Equal(t, scopePrefixes, authDetails.ScopePrefixes)
	})

	t.Run("returns nil when does not match", func(t *testing.T) {
		setup()
		delete(reqData.Body.Extra, identityAttributeKey)
		provider := tenantmapping.NewAuthenticatorContextProvider(nil, []authenticator.Config{})
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})

	t.Run("returns error during JWTAuthFlow when authenticator unique attribute is present but identity attribute is not specified in the Extra map of request body", func(t *testing.T) {
		setup()
		delete(reqData.Body.Extra, identityAttributeKey)
		provider := tenantmapping.NewAuthenticatorContextProvider(nil, authn)
		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.EqualError(t, err, apperrors.NewInvalidDataError(fmt.Sprintf("missing identity attribute from %q authenticator token", authn[0].Name)).Error())
	})
}
