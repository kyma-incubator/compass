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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/stretchr/testify/require"
)

const envPrefix = "APP"

func TestInitFromEnv(t *testing.T) {
	t.Run("When environment contains authenticator configuration", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name:          "TEST_AUTHN",
			ScopePrefixes: []string{"prefix!"},
			Attributes: authenticator.Attributes{
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
			},
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_SCOPE_PREFIXES", expectedAuthenticator.Name), expectedAuthenticator.ScopePrefixes[0])
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, 1, len(authenticators))
		require.Equal(t, expectedAuthenticator, authenticators[0])
	})

	t.Run("When environment contains authenticator configuration with multiple scope prefixes", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name:          "TEST_AUTHN",
			ScopePrefixes: []string{"prefix1!", "prefix2!"},
			Attributes: authenticator.Attributes{
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
			},
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_SCOPE_PREFIXES", expectedAuthenticator.Name), strings.Join(expectedAuthenticator.ScopePrefixes, ","))
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, 1, len(authenticators))
		require.Equal(t, expectedAuthenticator, authenticators[0])
	})

	t.Run("When environment contains multiple authenticator configurations", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator1 := authenticator.Config{
			Name:          "TEST_AUTHN1",
			ScopePrefixes: []string{"prefix!"},
			Attributes: authenticator.Attributes{
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
			},
		}

		expectedAuthenticator2 := authenticator.Config{
			Name:          "TEST_AUTHN2",
			ScopePrefixes: []string{"prefix!"},
			Attributes: authenticator.Attributes{
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
			},
		}

		expectedAuthenticators := []authenticator.Config{expectedAuthenticator1, expectedAuthenticator2}

		attributesJSON1, err := json.Marshal(expectedAuthenticator1.Attributes)
		require.NoError(t, err)

		attributesJSON2, err := json.Marshal(expectedAuthenticator2.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator1.Name), string(attributesJSON1))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_SCOPE_PREFIXES", expectedAuthenticator1.Name), expectedAuthenticator1.ScopePrefixes[0])
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator2.Name), string(attributesJSON2))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_SCOPE_PREFIXES", expectedAuthenticator2.Name), expectedAuthenticator2.ScopePrefixes[0])
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, len(expectedAuthenticators), len(authenticators))
		require.ElementsMatch(t, expectedAuthenticators, authenticators)
	})

	t.Run("When environment does not contain any authenticator configurations", func(t *testing.T) {
		os.Clearenv()

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, 0, len(authenticators))
	})

	t.Run("When environment contains authenticator configuration with invalid attribute", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name:          "TEST_AUTHN",
			ScopePrefixes: []string{"prefix!"},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   "test-unique-key",
					Value: "test-value",
				},
				TenantAttribute: authenticator.Attribute{
					Key: "",
				},
				IdentityAttribute: authenticator.Attribute{
					Key: "",
				},
			},
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_SCOPE_PREFIXES", expectedAuthenticator.Name), expectedAuthenticator.ScopePrefixes[0])
		require.NoError(t, err)

		_, err = authenticator.InitFromEnv(envPrefix)

		require.Error(t, err)
	})

	t.Run("When environment contains authenticator configuration with missing unique attribute value", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name:          "TEST_AUTHN",
			ScopePrefixes: []string{"prefix!"},
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   "test-unique-key",
					Value: "",
				},
				TenantAttribute: authenticator.Attribute{
					Key: "test-tenant-key",
				},
				IdentityAttribute: authenticator.Attribute{
					Key: "test-identity-key",
				},
			},
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_SCOPE_PREFIXES", expectedAuthenticator.Name), expectedAuthenticator.ScopePrefixes[0])
		require.NoError(t, err)

		_, err = authenticator.InitFromEnv(envPrefix)

		require.Error(t, err)
	})

	t.Run("When environment contains authenticator configuration with missing prefix should be okay", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name: "TEST_AUTHN",
			Attributes: authenticator.Attributes{
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
			},
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, 1, len(authenticators))
		require.Equal(t, expectedAuthenticator, authenticators[0])
	})
}
