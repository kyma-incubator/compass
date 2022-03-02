package director_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemAuthQuery(t *testing.T) {
	t.Run("Should return system auth query", func(t *testing.T) {
		// given
		authID := "someId"

		expectedQuery := `query {
		  result: systemAuth(id: "someId") {
			id
			auth {
			  certCommonName
			}
		  }
		}`

		expectedQuery = trimSpacesAndTabs(expectedQuery)

		// when
		query := trimSpacesAndTabs(director.SystemAuthQuery(authID))

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func TestSystemAuthByTokenQuery(t *testing.T) {
	t.Run("Should return system auth by token query", func(t *testing.T) {
		// given
		token := "qwerty"

		expectedQuery := `query {
		result: systemAuthByToken(token: "qwerty") {
			id
			auth {
				credential {
				  ... on BasicCredentialData {
					username
					password
				  }
				  ... on OAuthCredentialData {
					clientId
					clientSecret
					url
				  }
				}
				oneTimeToken {
				  __typename
				  token
				  used
				  expiresAt
				  connectorURL
				  rawEncoded
				  raw
				  usedAt
				  type
				  createdAt
				}
				certCommonName
				accessStrategy
				additionalHeaders
				additionalQueryParams
				requestAuth {
				  csrf {
					tokenEndpointURL
					credential {
					  ... on BasicCredentialData {
						username
						password
					  }
					  ... on OAuthCredentialData {
						clientId
						clientSecret
						url
					  }
					}
					additionalHeaders
					additionalQueryParams
				  }
				}
			  }
			}
		}`

		expectedQuery = trimSpacesAndTabs(expectedQuery)

		// when
		query := trimSpacesAndTabs(director.SystemAuthByTokenQuery(token))

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func TestTenantByExternalIDQuery(t *testing.T) {
	t.Run("Should return tenant by external ID query", func(t *testing.T) {
		// given
		tenantID := "b91b59f7-2563-40b2-aba9-fef726037aa3"

		expectedQuery := `query {
			result: tenantByExternalID(id: "b91b59f7-2563-40b2-aba9-fef726037aa3") {
				id
				internalID
				name
				type
				parentID
				labels
			}
		}`

		expectedQuery = trimSpacesAndTabs(expectedQuery)

		// when
		query := trimSpacesAndTabs(director.TenantByExternalIDQuery(tenantID))

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func TestTenantByInternalIDQuery(t *testing.T) {
	t.Run("Should return tenant by internal ID query", func(t *testing.T) {
		// given
		tenantID := "b91b59f7-2563-40b2-aba9-fef726037aa3"

		expectedQuery := `query {
			result: tenantByInternalID(id: "b91b59f7-2563-40b2-aba9-fef726037aa3") {
				id
				internalID
				name
				type
				parentID
				labels
			}
		}`

		expectedQuery = trimSpacesAndTabs(expectedQuery)

		// when
		query := trimSpacesAndTabs(director.TenantByInternalIDQuery(tenantID))

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func TestTenantByLowestOwnerForResourceQuery(t *testing.T) {
	t.Run("Should return tenant by lowest owner for resource query", func(t *testing.T) {
		// given
		resource := "runtime"
		id := "b91b59f7-2563-40b2-aba9-fef726037aa3"

		expectedQuery := `query {
			result: tenantByLowestOwnerForResource(id:"b91b59f7-2563-40b2-aba9-fef726037aa3", resource:"runtime")
		}`

		expectedQuery = trimSpacesAndTabs(expectedQuery)

		// when
		query := trimSpacesAndTabs(director.TenantByLowestOwnerForResourceQuery(id, resource))

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func TestUpdateSystemAuthQuery(t *testing.T) {
	t.Run("Should return tenant by lowest owner for resource query", func(t *testing.T) {
		// given
		auth := graphql.Auth{
			Credential: graphql.BasicCredentialData{
				Username: "user",
				Password: "pass",
			},
			AccessStrategy:                  nil,
			AdditionalHeaders:               nil,
			AdditionalHeadersSerialized:     nil,
			AdditionalQueryParams:           nil,
			AdditionalQueryParamsSerialized: nil,
			RequestAuth:                     nil,
			OneTimeToken:                    nil,
			CertCommonName:                  str.Ptr("CN"),
		}
		id := "b91b59f7-2563-40b2-aba9-fef726037aa3"

		expectedQuery := `query {
			result: tenantByLowestOwnerForResource(id:"b91b59f7-2563-40b2-aba9-fef726037aa3", resource:"runtime")
		}`

		expectedQuery = trimSpacesAndTabs(expectedQuery)

		// when
		rawQuery, err := director.UpdateSystemAuthQuery(id, auth)
		require.NoError(t, err)

		query := trimSpacesAndTabs(rawQuery)

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func trimSpacesAndTabs(str string) string {
	var res string

	res = strings.Replace(str, "\t", "", -1)
	res = strings.Replace(res, "\n", "", -1)

	return res
}
