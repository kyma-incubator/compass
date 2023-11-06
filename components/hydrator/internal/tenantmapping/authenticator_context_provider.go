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
	"math"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	directorErrors "github.com/kyma-incubator/compass/components/hydrator/internal/director"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/tidwall/gjson"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/pkg/errors"
)

// NewAuthenticatorContextProvider implements the ObjectContextProvider interface by looking for user scopes in the 'scope' token attribute
// and also extracts the tenant information from the token by using a dedicated TenantAttribute defined for the specified authenticator.
// It uses its authenticators to extract authentication details from the requestData.
func NewAuthenticatorContextProvider(clientProvider DirectorClient, authenticators []authenticator.Config) *authenticatorContextProvider {
	return &authenticatorContextProvider{
		directorClient: clientProvider,
		tenantKeys: KeysExtra{
			TenantKey:         tenantmapping.ConsumerTenantKey,
			ExternalTenantKey: tenantmapping.ExternalTenantKey,
		},
		authenticators: authenticators,
	}
}

type authenticatorContextProvider struct {
	directorClient DirectorClient
	tenantKeys     KeysExtra
	authenticators []authenticator.Config
}

// GetObjectContext is the authenticatorContextProvider implementation of the ObjectContextProvider interface
func (m *authenticatorContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	var scopes string

	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumer.User,
	})

	ctx = log.ContextWithLogger(ctx, logger)

	authn := authDetails.Authenticator

	log.C(ctx).Info("Getting scopes from token attribute")
	userScopes, err := reqData.GetUserScopes(authDetails.ScopePrefixes)
	if err != nil {
		return ObjectContext{}, err
	}
	scopes = strings.Join(userScopes, " ")

	extra, err := reqData.MarshalExtra()
	if err != nil {
		return ObjectContext{}, err
	}

	clientID := gjson.Get(extra, authn.Attributes.ClientID.Key).String()

	var externalTenantID string
	priority := math.MaxInt64
	for _, ta := range authn.Attributes.TenantsAttribute {
		currentTenant := gjson.Get(extra, ta.Key).String()
		if currentTenant == "" {
			log.C(ctx).Warnf("tenant attribute: %q missing from: %s authenticator token", ta.Key, authn.Name)
			continue
		}
		log.C(ctx).Infof("The priority of tenant: %s is: %d", currentTenant, ta.Priority)
		if ta.Priority < priority {
			priority = ta.Priority
			externalTenantID = currentTenant
		}
	}

	if externalTenantID == "" {
		return ObjectContext{}, errors.Errorf("missing tenant attribute from: %q authenticator token", authn.Name)
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)

	tenantMapping, err := m.directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(&graphql.Tenant{ID: externalTenantID}, m.tenantKeys, scopes, mergeWithOtherScopes, authDetails.Region, clientID, authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.AuthenticatorObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	objCtx := NewObjectContext(tenantMapping, m.tenantKeys, scopes, mergeWithOtherScopes, authDetails.Region, clientID, authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.AuthenticatorObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", RedactConsumerIDForLogging(objCtx))

	return objCtx, nil
}

// Match checks whether any of its preconfigured authenticators matches the ReqData and if so builds AuthDetails for the matched authenticator
func (m *authenticatorContextProvider) Match(ctx context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	coords, exist, err := data.ExtractCoordinates()
	if err != nil {
		return false, nil, errors.Wrap(err, "while extracting coordinates")
	}
	if exist {
		for _, authn := range m.authenticators {
			if authn.Name != coords.Name {
				continue
			}
			log.C(ctx).Infof("Request token matches %q authenticator", authn.Name)

			extra, err := data.MarshalExtra()
			if err != nil {
				return false, nil, err
			}

			authID := gjson.Get(extra, authn.Attributes.IdentityAttribute.Key).String()
			if len(authID) == 0 {
				return false, nil, apperrors.NewInvalidDataError("missing identity attribute from %q authenticator token", authn.Name)
			}

			index := coords.Index
			return true, &oathkeeper.AuthDetails{AuthID: authID, AuthFlow: oathkeeper.JWTAuthFlow, Authenticator: &authn, ScopePrefixes: authn.TrustedIssuers[index].ScopePrefixes, Region: authn.TrustedIssuers[index].Region}, nil
		}
	}

	return false, nil, nil
}
