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
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/pkg/errors"
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

	jwtAuthDetails := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow, ScopePrefix: scopePrefix}

	t.Run("returns tenant and scopes that are defined in the Extra map of ReqData", func(t *testing.T) {
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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
		require.Equal(t, strings.Join(expectedScopes, " "), objCtx.Scopes)
		require.Equal(t, username, objCtx.ConsumerID)
		require.Equal(t, userObjCtxType, string(objCtx.ConsumerType))
	})

	t.Run("returns tenant that is defined as a request header when it is missing from the the Extra map of ReqData", func(t *testing.T) {
		uniqueAttributeKey := "extra.unique"
		uniqueAttributeValue := "value"
		tenantAttributeKey := "tenant"
		header := http.Header{}
		header.Add(oathkeeper.ExternalTenantKey, expectedExternalTenantID.String())
		reqData := oathkeeper.ReqData{
			Header: header,
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ScopesKey: prefixedScopes,
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

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Equal(t, expectedTenantID.String(), objCtx.TenantID)
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

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(nil, missingTenantErr).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		objCtx, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.NoError(t, err)
		require.Empty(t, objCtx.TenantID)
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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

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

		tenantMappingModel := &model.BusinessTenantMapping{
			ID:             expectedTenantID.String(),
			ExternalTenant: expectedExternalTenantID.String(),
		}

		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(tenantMappingModel, nil).Once()

		authn := &authenticator.Config{
			Name: "test-authenticator",
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		_, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.EqualError(t, err, fmt.Sprintf("tenant attribute %q missing from %s authenticator token and %q request header", tenantAttributeKey, authn.Name, oathkeeper.ExternalTenantKey))
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
		tenantRepoMock := getTenantRepositoryMock()
		tenantRepoMock.On("GetByExternalTenant", mock.Anything, expectedExternalTenantID.String()).Return(nil, mockErr).Once()

		authn := &authenticator.Config{
			TrustedIssuers: []authenticator.TrustedIssuer{{
				ScopePrefix: scopePrefix,
			}},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
				TenantAttribute: authenticator.Attribute{
					Key: tenantAttributeKey,
				},
			},
		}

		userAuthDetailsWithAuthenticator := jwtAuthDetails
		userAuthDetailsWithAuthenticator.Authenticator = authn

		provider := tenantmapping.NewAuthenticatorContextProvider(tenantRepoMock)

		_, err := provider.GetObjectContext(context.TODO(), reqData, userAuthDetailsWithAuthenticator)

		require.EqualError(t, err, fmt.Sprintf("while getting external tenant mapping [ExternalTenantId=%s]: %s", expectedExternalTenantID, mockErr.Error()))
	})

}
