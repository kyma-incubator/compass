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
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/stretchr/testify/require"
)

const envPrefix = "APP"

func TestInitFromEnv(t *testing.T) {
	authenticatorDomain := "example.domain.com"
	authenticatorPrefix := "prefix!."
	authenticatorPrefixTrimmed := "prefix!"
	authenticator2Domain := "example2.domain.com"
	authenticator2Prefix := "prefix2!."
	authenticator2PrefixTrimmed := "prefix2!"
	trustedIssuers := []authenticator.TrustedIssuer{
		{
			DomainURL:   authenticatorDomain,
			ScopePrefix: authenticatorPrefix,
		},
		{
			DomainURL:   authenticator2Domain,
			ScopePrefix: authenticator2Prefix,
		},
	}
	clientIDSufixes := []string{authenticatorPrefixTrimmed, authenticator2PrefixTrimmed}

	t.Run("When environment contains authenticator configuration", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()
		expectedAuthenticator := authenticator.Config{
			Name:           "TEST_AUTHN",
			TrustedIssuers: trustedIssuers,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes,
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		issuersJSON, err := json.Marshal(trustedIssuers)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator.Name), string(issuersJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator.Name), "true")
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, 1, len(authenticators))
		require.Equal(t, expectedAuthenticator, authenticators[0])
	})

	t.Run("When environment contains multiple authenticator configurations", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()
		authenticatorDomain2 := "example.domain2.com"
		authenticatorPrefix2 := "prefix!2"
		trustedIssuers2 := []authenticator.TrustedIssuer{
			{
				DomainURL:   authenticatorDomain2,
				ScopePrefix: authenticatorPrefix2,
			},
		}
		clientIDSufixes2 := []string{authenticatorPrefix2}

		expectedAuthenticator1 := authenticator.Config{
			Name:           "TEST_AUTHN1",
			TrustedIssuers: trustedIssuers,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes,
		}

		expectedAuthenticator2 := authenticator.Config{
			Name:           "TEST_AUTHN2",
			TrustedIssuers: trustedIssuers2,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes2,
		}

		expectedAuthenticators := []authenticator.Config{expectedAuthenticator1, expectedAuthenticator2}

		attributesJSON1, err := json.Marshal(expectedAuthenticator1.Attributes)
		require.NoError(t, err)

		attributesJSON2, err := json.Marshal(expectedAuthenticator2.Attributes)
		require.NoError(t, err)

		issuersJSON1, err := json.Marshal(trustedIssuers)
		require.NoError(t, err)

		issuersJSON2, err := json.Marshal(trustedIssuers2)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator1.Name), string(attributesJSON1))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator1.Name), string(issuersJSON1))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator1.Name), "true")
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator2.Name), string(attributesJSON2))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator2.Name), string(issuersJSON2))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator2.Name), "true")
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, len(expectedAuthenticators), len(authenticators))
		require.ElementsMatch(t, expectedAuthenticators, authenticators)
	})

	t.Run("When environment contains multiple authenticator configurations - one has suffix configured and one does not ", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()
		authenticatorDomain2 := "example.domain2.com"
		authenticatorPrefix2 := "prefix!2."
		trustedIssuers2 := []authenticator.TrustedIssuer{
			{
				DomainURL:   authenticatorDomain2,
				ScopePrefix: authenticatorPrefix2,
			},
		}

		expectedAuthenticator1 := authenticator.Config{
			Name:           "TEST_AUTHN1",
			TrustedIssuers: trustedIssuers,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes,
		}

		expectedAuthenticator2 := authenticator.Config{
			Name:           "TEST_AUTHN2",
			TrustedIssuers: trustedIssuers2,
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

		issuersJSON1, err := json.Marshal(trustedIssuers)
		require.NoError(t, err)

		issuersJSON2, err := json.Marshal(trustedIssuers2)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator1.Name), string(attributesJSON1))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator1.Name), string(issuersJSON1))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator1.Name), "true")
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator2.Name), string(attributesJSON2))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator2.Name), string(issuersJSON2))
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
			Name:           "TEST_AUTHN",
			TrustedIssuers: trustedIssuers,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes,
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		issuersJSON, err := json.Marshal(trustedIssuers)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator.Name), string(issuersJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator.Name), "true")
		require.NoError(t, err)

		_, err = authenticator.InitFromEnv(envPrefix)

		require.Error(t, err)
	})

	t.Run("When environment contains authenticator configuration with missing unique attribute value", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name:           "TEST_AUTHN",
			TrustedIssuers: trustedIssuers,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes,
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		issuersJSON, err := json.Marshal(trustedIssuers)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator.Name), string(issuersJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator.Name), "true")
		require.NoError(t, err)

		_, err = authenticator.InitFromEnv(envPrefix)

		require.Error(t, err)
	})

	t.Run("When environment contains authenticator configuration with missing prefix should be okay", func(t *testing.T) {
		os.Clearenv()
		defer os.Clearenv()

		expectedAuthenticator := authenticator.Config{
			Name:           "TEST_AUTHN",
			TrustedIssuers: trustedIssuers,
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
			CheckSuffix:      true,
			ClientIDSuffixes: clientIDSufixes,
		}

		attributesJSON, err := json.Marshal(expectedAuthenticator.Attributes)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_ATTRIBUTES", expectedAuthenticator.Name), string(attributesJSON))
		require.NoError(t, err)

		issuersJSON, err := json.Marshal(trustedIssuers)
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_TRUSTED_ISSUERS", expectedAuthenticator.Name), string(issuersJSON))
		require.NoError(t, err)

		err = os.Setenv(fmt.Sprintf("APP_%s_AUTHENTICATOR_CHECK_CLIENT_ID_SUFFIX", expectedAuthenticator.Name), "true")
		require.NoError(t, err)

		authenticators, err := authenticator.InitFromEnv(envPrefix)

		require.NoError(t, err)
		require.Equal(t, 1, len(authenticators))
		require.Equal(t, expectedAuthenticator, authenticators[0])
	})
}
