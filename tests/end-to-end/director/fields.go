package director

import (
	"fmt"
	"strings"
)

// gqlFieldsProvider is responsible for generating GraphQL queries that request for all fields for given type
type gqlFieldsProvider struct{}

// fieldCtx is a map of optional fields that can be passed to FieldsProvider
// Map keys should be in following format: `type.field` eg. `APIDefinition.auth`
type fieldCtx map[string]string

// addFieldsFromContext checks if field context contains specific keys, adds them to provided fields and returns them
func addFieldsFromContext(oldFields string, ctx []fieldCtx, keys []string) string {
	var newFields []string
	for _, key := range keys {
		for _, dict := range ctx {
			if val, ok := dict[key]; ok {
				newFields = append(newFields, val)
				break
			}
		}
	}
	if len(newFields) == 0 {
		return oldFields
	}

	return fmt.Sprintf("%s\n%s", oldFields, strings.Join(newFields, "\n"))
}

func (fp *gqlFieldsProvider) Page(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, fp.ForPageInfo())
}

func (fp *gqlFieldsProvider) ForApplication(ctx ...fieldCtx) string {
	return fmt.Sprintf(`id
		name
		description
		labels
		status {condition timestamp}
		webhooks {%s}
		healthCheckURL
		apis {%s}
		eventAPIs {%s}
		documents {%s}
		auths {%s}
	`, fp.ForWebhooks(), fp.Page(fp.ForAPIDefinition(ctx...)), fp.Page(fp.ForEventAPI()), fp.Page(fp.ForDocument()), fp.ForSystemAuth())
}

func (fp *gqlFieldsProvider) ForWebhooks() string {
	return fmt.Sprintf(
		`id
		applicationID
		type
		url
		auth {
		  %s
		}`, fp.ForAuth())
}

func (fp *gqlFieldsProvider) ForAPIDefinition(ctx ...fieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`		id
		name
		description
		spec {%s}
		targetURL
		group
		auths {%s}
		defaultAuth {%s}
		version {%s}`, fp.ForApiSpec(), fp.ForAPIRuntimeAuth(), fp.ForAuth(), fp.ForVersion()),
		ctx, []string{"APIDefinition.auth"})
}

func (fp *gqlFieldsProvider) ForSystemAuth() string {
	return fmt.Sprintf(`
		id
		auth {%s}`, fp.ForAuth())
}

func (fp *gqlFieldsProvider) ForApiSpec() string {
	return fmt.Sprintf(`data
		format
		type
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *gqlFieldsProvider) ForFetchRequest() string {
	return fmt.Sprintf(`url
		auth {%s}
		mode
		filter
		status {condition timestamp}`, fp.ForAuth())
}

func (fp *gqlFieldsProvider) ForAPIRuntimeAuth() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, fp.ForAuth())
}

func (fp *gqlFieldsProvider) ForVersion() string {
	return `value
		deprecated
		deprecatedSince
		forRemoval`
}

func (fp *gqlFieldsProvider) ForPageInfo() string {
	return `startCursor
		endCursor
		hasNextPage`
}

func (fp *gqlFieldsProvider) ForEventAPI() string {
	return fmt.Sprintf(`
			id
			applicationID
			name
			description
			group 
			spec {%s}
			version {%s}
		`, fp.ForEventSpec(), fp.ForVersion())
}

func (fp *gqlFieldsProvider) ForEventSpec() string {
	return fmt.Sprintf(`data
		type
		format
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *gqlFieldsProvider) ForDocument() string {
	return fmt.Sprintf(`
		id
		applicationID
		title
		displayName
		description
		format
		kind
		data
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *gqlFieldsProvider) ForAuth() string {
	return fmt.Sprintf(`credential {
				... on BasicCredentialData {
					username
					password
				}
				...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				}
			}
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
				  ...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				  }
			    }
				additionalHeaders
				additionalQueryParams
			}
			}
		`)
}

func (fp *gqlFieldsProvider) ForLabel() string {
	return `key
			value`
}

func (fp *gqlFieldsProvider) ForRuntime() string {
	return fmt.Sprintf(`
		id
		name
		description
		labels 
		status {condition timestamp}
		auths {%s}`, fp.ForSystemAuth())
}

func (fp *gqlFieldsProvider) ForApplicationLabel() string {
	return `
		key
		value`
}

func (fp *gqlFieldsProvider) ForLabelDefinition() string {
	return `
		key
		schema`
}

func (fp *gqlFieldsProvider) ForOneTimeToken() string {
	return `
		token
		connectorURL`
}
