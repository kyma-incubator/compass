package graphqlizer

import (
	"fmt"
	"strings"
)

// GqlFieldsProvider is responsible for generating GraphQL queries that request for all fields for given type
type GqlFieldsProvider struct{}

// fieldCtx is a map of optional fields that can be passed to FieldsProvider
// Map keys should be in following format: `type.field` eg. `APIDefinition.auth`
type FieldCtx map[string]string

// addFieldsFromContext checks if field context contains specific keys, adds them to provided fields and returns them
func addFieldsFromContext(oldFields string, ctx []FieldCtx, keys []string) string {
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

func (fp *GqlFieldsProvider) Page(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, fp.ForPageInfo())
}

func (fp *GqlFieldsProvider) ForApplication(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		providerName
		description
		integrationSystemID
		labels
		status {condition timestamp}
		webhooks {%s}
		healthCheckURL
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}
		packages {%s}
		auths {%s}
		eventingConfiguration { defaultURL }
	`, fp.ForWebhooks(), fp.Page(fp.ForAPIDefinition(ctx...)), fp.Page(fp.ForEventDefinition()), fp.Page(fp.ForDocument()), fp.Page(fp.ForPackage()), fp.ForSystemAuth()),
		ctx, []string{"Application.package", "Application.apiDefinition", "Application.eventDefinition"})
}

func (fp *GqlFieldsProvider) ForApplicationTemplate(ctx ...FieldCtx) string {
	return fmt.Sprintf(`id
		name
		description
		applicationInput
		placeholders {%s}
		accessLevel
	`, fp.ForPlaceholders())
}

func (fp *GqlFieldsProvider) ForWebhooks() string {
	return fmt.Sprintf(
		`id
		applicationID
		type
		url
		auth {
		  %s
		}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) ForAPIDefinition(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`id
		name
		applicationID
		description
		spec {%s}
		targetURL
		group
		auths {%s}
		defaultAuth {%s}
		version {%s}`, fp.ForApiSpec(), fp.ForAPIRuntimeAuth(), fp.ForAuth(), fp.ForVersion()),
		ctx, []string{"APIDefinition.auth"})
}

func (fp *GqlFieldsProvider) ForSystemAuth() string {
	return fmt.Sprintf(`
		id
		auth {%s}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) ForApiSpec() string {
	return fmt.Sprintf(`data
		format
		type
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *GqlFieldsProvider) ForFetchRequest() string {
	return fmt.Sprintf(`url
		auth {%s}
		mode
		filter
		status {condition timestamp}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) ForAPIRuntimeAuth() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) ForVersion() string {
	return `value
		deprecated
		deprecatedSince
		forRemoval`
}

func (fp *GqlFieldsProvider) ForPageInfo() string {
	return `startCursor
		endCursor
		hasNextPage`
}

func (fp *GqlFieldsProvider) ForEventDefinition() string {
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

func (fp *GqlFieldsProvider) ForEventSpec() string {
	return fmt.Sprintf(`data
		type
		format
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *GqlFieldsProvider) ForDocument() string {
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

func (fp *GqlFieldsProvider) ForAuth() string {
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

func (fp *GqlFieldsProvider) ForLabel() string {
	return `key
			value`
}

func (fp *GqlFieldsProvider) ForRuntime() string {
	return fmt.Sprintf(`
		id
		name
		description
		labels 
		status {condition timestamp}
		metadata { creationTimestamp }
		auths {%s}
		eventingConfiguration { defaultURL }`, fp.ForSystemAuth())
}

func (fp *GqlFieldsProvider) ForApplicationLabel() string {
	return `
		key
		value`
}

func (fp *GqlFieldsProvider) ForLabelDefinition() string {
	return `
		key
		schema`
}

func (fp *GqlFieldsProvider) ForOneTimeTokenForApplication() string {
	return `
		token
		connectorURL
		raw
		rawEncoded
		legacyConnectorURL`
}

func (fp *GqlFieldsProvider) ForOneTimeTokenForRuntime() string {
	return `
		token
		connectorURL
		raw
		rawEncoded`
}

func (fp *GqlFieldsProvider) ForIntegrationSystem() string {
	return fmt.Sprintf(`
		id
		name
		description
		auths {%s}`, fp.ForSystemAuth())
}

func (fp *GqlFieldsProvider) ForPlaceholders() string {
	return `
		name
		description`
}

func (fp *GqlFieldsProvider) ForEventingConfiguration() string {
	return `
		defaultURL`
}

func (fp *GqlFieldsProvider) ForViewer() string {
	return `
		id
		type`
}

func (fp *GqlFieldsProvider) ForTenant() string {
	return `
		id
		name`
}

func (fp *GqlFieldsProvider) ForPackage(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		description
		instanceAuthRequestInputSchema
		instanceAuths {%s}
		defaultInstanceAuth {%s}
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}`, fp.ForPackageInstanceAuth(), fp.ForAuth(), fp.Page(fp.ForAPIDefinition(ctx...)), fp.Page(fp.ForEventDefinition()), fp.Page(fp.ForDocument())),
		ctx, []string{"Package.instanceAuth"})
}

func (fp *GqlFieldsProvider) ForPackageInstanceAuth() string {
	return fmt.Sprintf(`
		id
		context
		inputParams
		auth {%s}
		status {%s}`, fp.ForAuth(), fp.ForPackageInstanceAuthStatus())
}

func (fp *GqlFieldsProvider) ForPackageInstanceAuthStatus() string {
	return `
		condition
		timestamp
		message
		reason`
}

func (fp *GqlFieldsProvider) ForAutomaticScenarioAssignment() string {
	return fmt.Sprintf(`
		scenarioName
		selector {%s}`, fp.ForLabel())
}
