package director

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
)

func SystemAuthQuery(authID string) string {
	return fmt.Sprintf(`query {
	  	result: systemAuth(id: "%s") {
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
		}`, authID)
}

func SystemAuthByTokenQuery(token string) string {
	return fmt.Sprintf(`query {
		result: systemAuthByToken(token: "%s") {
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
		}`, token)
}

func TenantByExternalIDQuery(tenantID string) string {
	return fmt.Sprintf(`query {
		result: tenantByExternalID(id: "%s") {
			id
			internalID
			name
			type
			parentID
			labels
		}
	}`, tenantID)
}

func TenantByInternalIDQuery(tenantID string) string {
	return fmt.Sprintf(`query {
		result: tenantByInternalID(id: "%s") {
			id
			internalID
			name
			type
			parentID
			labels
		}
	}`, tenantID)
}

func TenantByLowestOwnerForResourceQuery(resourceID, resourceType string) string {
	return fmt.Sprintf(`query {
		result: tenantByLowestOwnerForResource(id:"%s", resource:"%s")
	}`, resourceID, resourceType)
}

func UpdateSystemAuthQuery(sysAuth *model.SystemAuth) (string, error) {
	authInput, err := auth.ToGraphQLInput(sysAuth.Value)
	if err != nil {
		return "", err
	}

	g := graphqlizer.Graphqlizer{}
	gqlAuthInput, err := g.AuthInputToGQL(authInput)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`mutation {
		result: updateSystemAuth(authID: "%s", in: %s) { id }
	}`, sysAuth.ID, gqlAuthInput), nil
}

func InvalidateSystemAuthOneTimeTokenQuery(authID string) string {
	return fmt.Sprintf(`mutation {
		result: invalidateSystemAuthOneTimeToken(authID: "%s") {
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
		}`, authID)
}

func RuntimeByTokenIssuerQuery(issuer string) string {
	return fmt.Sprintf(`query {
		result: runtimeByTokenIssuer(issuer: "%s") {
			id
    		name
	  	}
	}`, issuer)
}
