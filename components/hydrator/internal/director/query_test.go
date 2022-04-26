package director_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestSystemAuthQuery(t *testing.T) {
	t.Run("Should return system auth query", func(t *testing.T) {
		// given
		authID := "someId"

		expectedQuery := `query {
		result: systemAuth(id: "someId") {
			id
			tenantId
			referenceObjectId
			type
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

		expectedQuery = trimTabsAndNewLineSpace(expectedQuery)

		// when
		query := trimTabsAndNewLineSpace(director.SystemAuthQuery(authID))

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
			tenantId
			referenceObjectId
			type
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

		expectedQuery = trimTabsAndNewLineSpace(expectedQuery)

		// when
		query := trimTabsAndNewLineSpace(director.SystemAuthByTokenQuery(token))

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

		expectedQuery = trimTabsAndNewLineSpace(expectedQuery)

		// when
		query := trimTabsAndNewLineSpace(director.TenantByExternalIDQuery(tenantID))

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

		expectedQuery = trimTabsAndNewLineSpace(expectedQuery)

		// when
		query := trimTabsAndNewLineSpace(director.TenantByInternalIDQuery(tenantID))

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

		expectedQuery = trimTabsAndNewLineSpace(expectedQuery)

		// when
		query := trimTabsAndNewLineSpace(director.TenantByLowestOwnerForResourceQuery(id, resource))

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func TestUpdateSystemAuthQuery(t *testing.T) {
	t.Run("Should return update system auth query", func(t *testing.T) {
		// given
		authID := "b91b59f7-2563-40b2-aba9-fef726037aa3"
		refObjID := uuid.New()
		expectedTenantID := uuid.New()

		authData := &graphql.Auth{
			OneTimeToken:   nil,
			CertCommonName: str.Ptr("CN"),
			Credential: &graphql.BasicCredentialData{
				Username: "user",
				Password: "pass",
			},
		}

		authDataModel, err := auth.ToModel(authData)
		require.NoError(t, err)

		sysAuth := &model.SystemAuth{
			ID:       authID,
			TenantID: str.Ptr(expectedTenantID.String()),
			AppID:    str.Ptr(refObjID.String()),
			Value:    authDataModel,
		}

		expectedQuery := `mutation {
			result: updateSystemAuth(authID: "b91b59f7-2563-40b2-aba9-fef726037aa3", in: {credential:  {basic: {username: "user",password: "pass",},},}) { id }}`

		expectedQuery = trimTabsAndNewLineSpace(expectedQuery)

		// when
		rawQuery, err := director.UpdateSystemAuthQuery(sysAuth)
		require.NoError(t, err)

		query := trimTabsAndNewLineSpace(rawQuery)

		// then
		assert.Equal(t, expectedQuery, query)
	})
}

func trimTabsAndNewLineSpace(str string) string {
	var res string

	res = strings.Replace(str, "\t", "", -1)
	res = strings.Replace(res, "\n", "", -1)
	res = strings.Replace(res, " ", "", -1)

	return res
}
