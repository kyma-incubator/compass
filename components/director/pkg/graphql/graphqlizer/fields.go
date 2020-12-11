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

func buildProperties(allProperties map[string]string, omittedProperties []string) string {
	resultProperties := make([]string, 0, len(allProperties))
	for key, prop := range allProperties {
		shouldBeOmitted := false
		for _, om := range omittedProperties {
			if strings.HasPrefix(key, om) {
				shouldBeOmitted = true
				break
			}
		}
		if !shouldBeOmitted {
			resultProperties = append(resultProperties, prop)
		}
	}
	return strings.Join(resultProperties, "\n")
}

func extractOmitFor(omittedProperties []string, propertyName string) []string {
	result := make([]string, 0)
	for _, om := range omittedProperties {
		if strings.HasPrefix(om, propertyName+".") {
			result = append(result, strings.TrimPrefix(om, propertyName+"."))
		}
	}
	return result
}

func (fp *GqlFieldsProvider) Page(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, fp.ForPageInfo())
}

func (fp *GqlFieldsProvider) OmitForApplication(omittedProperties []string) string {
	packagesOmittedProperties := extractOmitFor(omittedProperties, "packages")
	webhooksOmittedProperties := extractOmitFor(omittedProperties, "webhooks")

	return buildProperties(map[string]string{
		"id":                    "id",
		"name":                  "name",
		"providerName":          "providerName",
		"description":           "description",
		"integrationSystemID":   "integrationSystemID",
		"labels":                "labels",
		"status":                "status { condition timestamp }",
		"webhooks":              fmt.Sprintf("webhooks {%s}", fp.OmitForWebhooks(webhooksOmittedProperties)),
		"healthCheckURL":        "healthCheckURL",
		"packages":              fmt.Sprintf("packages {%s}", fp.Page(fp.OmitForPackage(packagesOmittedProperties))),
		"auths":                 fmt.Sprintf("auths {%s}", fp.ForSystemAuth()),
		"eventingConfiguration": "eventingConfiguration { defaultURL }",
	}, omittedProperties)
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
		packages {%s}
		auths {%s}
		eventingConfiguration { defaultURL }
	`, fp.ForWebhooks(), fp.Page(fp.ForPackage()), fp.ForSystemAuth()),
		ctx, []string{"Application.package", "Application.apiDefinition", "Application.eventDefinition"})
}

func (fp *GqlFieldsProvider) ForApplicationTemplate(ctx ...FieldCtx) string {
	return fmt.Sprintf(`
		id
		name
		description
		applicationInput
		placeholders {%s}
		accessLevel
	`, fp.ForPlaceholders())
}

func (fp *GqlFieldsProvider) OmitForWebhooks(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"id":            "id",
		"applicationID": "applicationID",
		"type":          "type",
		"url":           "url",
		"auth":          fmt.Sprintf("auth {%s}", fp.ForAuth()),
	}, omittedProperties)
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

func (fp *GqlFieldsProvider) OmitForAPIDefinition(omittedProperties []string) string {
	specOmittedProperties := extractOmitFor(omittedProperties, "spec")
	versionOmittedProperties := extractOmitFor(omittedProperties, "version")

	return buildProperties(map[string]string{
		"id":          "id",
		"name":        "name",
		"description": "description",
		"spec":        fmt.Sprintf("spec {%s}", fp.OmitForApiSpec(specOmittedProperties)),
		"targetURL":   "targetURL",
		"group":       "group",
		"version":     fmt.Sprintf("version {%s}", fp.OmitForVersion(versionOmittedProperties)),
	}, omittedProperties)
}

func (fp *GqlFieldsProvider) ForAPIDefinition(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		description
		spec {%s}
		targetURL
		group
		version {%s}`, fp.ForApiSpec(), fp.ForVersion()),
		ctx, []string{"APIDefinition.auth"})
}

func (fp *GqlFieldsProvider) ForSystemAuth() string {
	return fmt.Sprintf(`
		id
		auth {%s}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) OmitForApiSpec(omittedProperties []string) string {
	frOmittedProperties := extractOmitFor(omittedProperties, "fetchRequest")

	return buildProperties(map[string]string{
		"data":         "data",
		"format":       "format",
		"type":         "type",
		"fetchRequest": fmt.Sprintf("fetchRequest {%s}", fp.OmitForFetchRequest(frOmittedProperties)),
	}, omittedProperties)
}

func (fp *GqlFieldsProvider) ForApiSpec() string {
	return fmt.Sprintf(`
		data
		format
		type
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *GqlFieldsProvider) OmitForFetchRequest(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"url":    "url",
		"auth":   fmt.Sprintf("auth {%s}", fp.ForAuth()),
		"mode":   "mode",
		"filter": "filter",
		"status": "status {condition message timestamp}",
	}, omittedProperties)
}

func (fp *GqlFieldsProvider) ForFetchRequest() string {
	return fmt.Sprintf(`
		url
		auth {%s}
		mode
		filter
		status {condition message timestamp}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) ForAPIRuntimeAuth() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, fp.ForAuth())
}

func (fp *GqlFieldsProvider) OmitForVersion(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"value":           "value",
		"deprecated":      "deprecated",
		"deprecatedSince": "deprecatedSince",
		"forRemoval":      "forRemoval",
	}, omittedProperties)
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

func (fp *GqlFieldsProvider) OmitForEventDefinition(omittedProperties []string) string {
	specOmittedProperties := extractOmitFor(omittedProperties, "spec")
	versionOmittedProperties := extractOmitFor(omittedProperties, "version")

	return buildProperties(map[string]string{
		"id":          "id",
		"name":        "name",
		"description": "description",
		"group":       "group",
		"spec":        fmt.Sprintf("spec {%s}", fp.OmitForEventSpec(specOmittedProperties)),
		"version":     fmt.Sprintf("version {%s}", fp.OmitForVersion(versionOmittedProperties)),
	}, omittedProperties)

}

func (fp *GqlFieldsProvider) ForEventDefinition() string {
	return fmt.Sprintf(`
			id
			name
			description
			group
			spec {%s}
			version {%s}
		`, fp.ForEventSpec(), fp.ForVersion())
}

func (fp *GqlFieldsProvider) OmitForEventSpec(omittedProperties []string) string {
	frOmittedProperties := extractOmitFor(omittedProperties, "fetchRequest")

	return buildProperties(map[string]string{
		"data":         "data",
		"type":         "type",
		"format":       "format",
		"fetchRequest": fmt.Sprintf("fetchRequest {%s}", fp.OmitForFetchRequest(frOmittedProperties)),
	}, omittedProperties)
}

func (fp *GqlFieldsProvider) ForEventSpec() string {
	return fmt.Sprintf(`
		data
		type
		format
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *GqlFieldsProvider) OmitForDocument(omittedProperties []string) string {
	frOmittedProperties := extractOmitFor(omittedProperties, "fetchRequest")

	return buildProperties(map[string]string{
		"id":           "id",
		"title":        "title",
		"displayName":  "displayName",
		"description":  "description",
		"format":       "format",
		"kind":         "kind",
		"data":         "data",
		"fetchRequest": fmt.Sprintf("fetchRequest {%s}", fp.OmitForFetchRequest(frOmittedProperties)),
	}, omittedProperties)
}

func (fp *GqlFieldsProvider) ForDocument() string {
	return fmt.Sprintf(`
		id
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
		internalID
		name
		initialized`
}

func (fp *GqlFieldsProvider) OmitForPackage(omittedProperties []string) string {
	apiDefOmittedProperties := extractOmitFor(omittedProperties, "apiDefinitions")
	eventDefOmittedProperties := extractOmitFor(omittedProperties, "eventDefinitions")
	documentsOmittedProperties := extractOmitFor(omittedProperties, "documents")
	instanceAuthsOmittedProperties := extractOmitFor(omittedProperties, "instanceAuths")

	return buildProperties(map[string]string{
		"id":                             "id",
		"name":                           "name",
		"description":                    "description",
		"instanceAuthRequestInputSchema": "instanceAuthRequestInputSchema",
		"instanceAuths":                  fmt.Sprintf("instanceAuths {%s}", fp.OmitForPackageInstanceAuth(instanceAuthsOmittedProperties)),
		"defaultInstanceAuth":            fmt.Sprintf("defaultInstanceAuth {%s}", fp.ForAuth()),
		"apiDefinitions":                 fmt.Sprintf("apiDefinitions {%s}", fp.Page(fp.OmitForAPIDefinition(apiDefOmittedProperties))),
		"eventDefinitions":               fmt.Sprintf("eventDefinitions {%s}", fp.Page(fp.OmitForEventDefinition(eventDefOmittedProperties))),
		"documents":                      fmt.Sprintf("documents {%s}", fp.Page(fp.OmitForDocument(documentsOmittedProperties))),
	}, omittedProperties)
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

func (fp *GqlFieldsProvider) OmitForPackageInstanceAuth(omittedProperties []string) string {
	statusOmittedProperties := extractOmitFor(omittedProperties, "status")

	return buildProperties(map[string]string{
		"id":          "id",
		"context":     "context",
		"inputParams": "inputParams",
		"auth":        fmt.Sprintf("auth {%s}", fp.ForAuth()),
		"status":      fmt.Sprintf("status {%s}", fp.OmitForPackageInstanceAuthStatus(statusOmittedProperties)),
	}, omittedProperties)
}

func (fp *GqlFieldsProvider) ForPackageInstanceAuth() string {
	return fmt.Sprintf(`
		id
		context
		inputParams
		auth {%s}
		status {%s}`, fp.ForAuth(), fp.ForPackageInstanceAuthStatus())
}

func (fp *GqlFieldsProvider) OmitForPackageInstanceAuthStatus(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"condition": "condition",
		"timestamp": "timestamp",
		"message":   "message",
		"reason":    "reason",
	}, omittedProperties)
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
