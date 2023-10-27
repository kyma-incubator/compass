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

package authenticator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"
)

func TestValidateAttribute(t *testing.T) {
	testCases := []struct {
		name      string
		attribute authenticator.Attribute
		errorMsg  string
	}{
		{
			name: "Success when attribute key is valid",
			attribute: authenticator.Attribute{
				Key:   "test-key",
				Value: "test-value",
			},
		},
		{
			name: "Error when attribute key is empty",
			attribute: authenticator.Attribute{
				Key: "",
			},
			errorMsg: "attribute key cannot be empty",
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			err := ts.attribute.Validate()
			if ts.errorMsg != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), ts.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAttributes(t *testing.T) {
	testUniqueKey := "test-unique-key"
	testUniqueValue := "test-unique-value"
	testIdentityKey := "test-identity-key"

	testCases := []struct {
		name       string
		attributes authenticator.Attributes
		errorMsg   string
	}{
		{
			name: "Success when all attributes are valid",
			attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   testUniqueKey,
					Value: testUniqueValue,
				},
				IdentityAttribute: authenticator.Attribute{
					Key: testIdentityKey,
				},
			},
		},
		{
			name: "Error when an attribute is invalid",
			attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   testUniqueKey,
					Value: testUniqueValue,
				},
				IdentityAttribute: authenticator.Attribute{},
			},
			errorMsg: "attribute key cannot be empty",
		},
		{
			name: "Error when a tenant attribute is invalid",
			attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   testUniqueKey,
					Value: testUniqueValue,
				},
				IdentityAttribute: authenticator.Attribute{
					Key: testIdentityKey,
				},
				TenantsAttribute: []authenticator.TenantAttribute{{}},
			},
			errorMsg: "tenant attribute key cannot be empty",
		},
		{
			name: "Error when the unique attribute value is missing",
			attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key: testUniqueKey,
				},
				IdentityAttribute: authenticator.Attribute{
					Key: testIdentityKey,
				},
			},
			errorMsg: "unique attribute value cannot be empty",
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			err := ts.attributes.Validate()
			if ts.errorMsg != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), ts.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateTenantAttribute(t *testing.T) {
	tenantAttributeKey := "testTenantAttributeKey"
	tenantAttributeKey2 := "testTenantAttributeKey2"

	testCases := []struct {
		name             string
		tenantAttributes []authenticator.TenantAttribute
		errorMsg         string
	}{
		{
			name:             "Success with one tenant attribute",
			tenantAttributes: []authenticator.TenantAttribute{{Key: tenantAttributeKey}},
		},
		{
			name:             "Success with multiple tenant attributes",
			tenantAttributes: []authenticator.TenantAttribute{{Key: tenantAttributeKey, Priority: 1}, {Key: tenantAttributeKey2, Priority: 2}},
		},
		{
			name:             "Error when key property is missing from the tenant attributes",
			tenantAttributes: []authenticator.TenantAttribute{{}},
			errorMsg:         "tenant attribute key cannot be empty",
		},
		{
			name:             "Error when tenant attribute priorities are incorrect",
			tenantAttributes: []authenticator.TenantAttribute{{Key: tenantAttributeKey, Priority: 1}, {Key: tenantAttributeKey2, Priority: 1}},
			errorMsg:         "tenant attribute priorities should be provided and should have unique values for each of them",
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			err := authenticator.Validate(ts.tenantAttributes)
			if ts.errorMsg != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), ts.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
