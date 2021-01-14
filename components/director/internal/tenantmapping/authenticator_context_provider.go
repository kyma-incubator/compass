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

package tenantmapping

import (
	"context"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// NewAuthenticatorContextProvider implements the ObjectContextProvider interface by looking for user scopes in the 'scope' token attribute
// and also extracts the tenant information from the token by using a dedicated TenantAttribute defined for the specified authenticator.
func NewAuthenticatorContextProvider(tenantRepo TenantRepository) *authenticatorContextProvider {
	return &authenticatorContextProvider{
		tenantRepo: tenantRepo,
	}
}

type authenticatorContextProvider struct {
	tenantRepo TenantRepository
}

// GetObjectContext is the authenticatorContextProvider implementation of the ObjectContextProvider interface
func (m *authenticatorContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	var externalTenantID, scopes string

	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumer.User,
	})

	ctx = log.ContextWithLogger(ctx, logger)

	authn := authDetails.Authenticator

	log.C(ctx).Info("Getting scopes from token attribute")
	userScopes, err := reqData.GetUserScopes(authDetails.ScopePrefix)
	if err != nil {
		return ObjectContext{}, err
	}
	scopes = strings.Join(userScopes, " ")

	extra, err := reqData.MarshalExtra()
	if err != nil {
		return ObjectContext{}, err
	}

	externalTenantID = tenantID(extra, authn.Attributes.TenantAttribute.Key, reqData.Header)
	if externalTenantID == "" {
		return ObjectContext{}, errors.Errorf("tenant attribute %q missing from %s authenticator token and %q request header", authn.Attributes.TenantAttribute.Key, authn.Name, oathkeeper.ExternalTenantKey)
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), scopes, authDetails.AuthID, consumer.User), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, authDetails.AuthID, consumer.User)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	return objCtx, nil
}

func tenantID(extra string, tenantKeyInExtra string, reqHeaders http.Header) string {
	id := gjson.Get(extra, tenantKeyInExtra).String()
	if id == "" {
		id = reqHeaders.Get(oathkeeper.ExternalTenantKey)
	}

	return id
}
