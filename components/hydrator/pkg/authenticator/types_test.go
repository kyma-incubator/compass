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

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/stretchr/testify/assert"
)

func TestValidateAttribute(t *testing.T) {
	t.Run("When key is valid", func(t *testing.T) {
		attribute := authenticator.Attribute{
			Key:   "test-key",
			Value: "test-value",
		}

		// when
		err := attribute.Validate()

		// then
		assert.NoError(t, err)
	})

	t.Run("When key is empty", func(t *testing.T) {
		attribute := authenticator.Attribute{
			Key:   "",
			Value: "",
		}

		// when
		err := attribute.Validate()

		// then
		assert.Error(t, err)
	})
}

func TestValidateAttributes(t *testing.T) {
	t.Run("When all attributes are valid", func(t *testing.T) {
		attributes := authenticator.Attributes{
			UniqueAttribute: authenticator.Attribute{
				Key:   "test-unique-key",
				Value: "test-value",
			},
			TenantAttribute: authenticator.Attribute{
				Key: "test-tenant-key",
			},
			IdentityAttribute: authenticator.Attribute{
				Key: "test-identity-key",
			},
		}

		// when
		err := attributes.Validate()

		// then
		assert.NoError(t, err)
	})

	t.Run("When an attribute is invalid", func(t *testing.T) {
		attributes := authenticator.Attributes{
			UniqueAttribute: authenticator.Attribute{
				Key:   "test-unique-key",
				Value: "test-value",
			},
			TenantAttribute: authenticator.Attribute{
				Key: "",
			},
			IdentityAttribute: authenticator.Attribute{
				Key: "test-identity-key",
			},
		}

		// when
		err := attributes.Validate()

		// then
		assert.Error(t, err)
	})

	t.Run("When the unique attribute value is missing", func(t *testing.T) {
		attributes := authenticator.Attributes{
			UniqueAttribute: authenticator.Attribute{
				Key: "test-unique-key",
			},
			TenantAttribute: authenticator.Attribute{
				Key: "test-tenant-key",
			},
			IdentityAttribute: authenticator.Attribute{
				Key: "test-identity-key",
			},
		}

		// when
		err := attributes.Validate()

		// then
		assert.Error(t, err)
	})
}
