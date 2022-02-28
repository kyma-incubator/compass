package director

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
)

func systemAuthQuery(authID string) string {
	return fmt.Sprintf(`query {
	  result: systemAuth(id: "%s") {
		id
		auth {
		  certCommonName
		}
	  }
	}`, authID)
}

func systemAuthByTokenQuery(token string) string {
	return fmt.Sprintf(`query {
		result: systemAuthByToken(token: "%s") {
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
		}`, token)
}

func tenantByExternalIDQuery(tenantID string) string {
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

func tenantByInternalIDQuery(tenantID string) string {
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

func tenantByLowestOwnerForResourceQuery(resourceID, resourceType string) string {
	return fmt.Sprintf(`query {
	  result: tenantByLowestOwnerForResource(id:"%s", resource:"%s")
	}`, resourceID, resourceType)
}

func updateSystemAuthQuery(authID string, gqlAuth graphql.Auth) (string, error) {
	authInput, err := auth.ToGraphQLInput(gqlAuth)
	if err != nil {
		return "", err
	}

	g := graphqlizer.Graphqlizer{}
	gqlAuthInput, err := g.AuthInputToGQL(authInput)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`mutation {
	  	result: updateSystemAuth(authID: "%s", in: %s) {
			id
	  	}
	}`, authID, gqlAuthInput), nil
}

func invalidateSystemAuthOneTimeTokenQuery(authID string) string {
	return fmt.Sprintf(`query {
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

func runtimeByTokenIssuerQuery(issuer string) string {
	return fmt.Sprintf(`query {
		result: runtimeByTokenIssuer(issuer: "%s") {
			id
    		name
	  	}
	}`, issuer)
}
