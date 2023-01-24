package graphqlizer

import (
	"fmt"
	"strings"
)

// GqlFieldsProvider is responsible for generating GraphQL queries that request for all fields for given type
type GqlFieldsProvider struct{}

// FieldCtx is a map of optional fields that can be passed to FieldsProvider
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

// Page missing godoc
func (fp *GqlFieldsProvider) Page(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, fp.ForPageInfo())
}

// OmitForApplication missing godoc
func (fp *GqlFieldsProvider) OmitForApplication(omittedProperties []string) string {
	bundlesOmittedProperties := extractOmitFor(omittedProperties, "bundles")
	webhooksOmittedProperties := extractOmitFor(omittedProperties, "webhooks")

	return buildProperties(map[string]string{
		"id":                    "id",
		"name":                  "name",
		"providerName":          "providerName",
		"description":           "description",
		"baseUrl":               "baseUrl",
		"systemNumber":          "systemNumber",
		"integrationSystemID":   "integrationSystemID",
		"applicationTemplateID": "applicationTemplateID",
		"labels":                "labels",
		"status":                "status { condition timestamp }",
		"webhooks":              fmt.Sprintf("webhooks {%s}", fp.OmitForWebhooks(webhooksOmittedProperties)),
		"healthCheckURL":        "healthCheckURL",
		"bundles":               fmt.Sprintf("bundles {%s}", fp.Page(fp.OmitForBundle(bundlesOmittedProperties))),
		"auths":                 fmt.Sprintf("auths {%s}", fp.ForSystemAuth()),
		"eventingConfiguration": "eventingConfiguration { defaultURL }",
	}, omittedProperties)
}

// ForApplication missing godoc
func (fp *GqlFieldsProvider) ForApplication(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		providerName
		description
		baseUrl
		systemNumber
		systemStatus
		integrationSystemID
		applicationTemplateID
		labels
		deletedAt
		updatedAt
		error
		status {condition timestamp}
		webhooks {%s}
		healthCheckURL
		bundles {%s}
		auths {%s}
		eventingConfiguration { defaultURL }
	`, fp.ForWebhooks(), fp.Page(fp.ForBundle()), fp.ForSystemAuth()),
		ctx, []string{"Application.bundle", "Application.apiDefinition", "Application.eventDefinition"})
}

// ForApplicationTemplate missing godoc
func (fp *GqlFieldsProvider) ForApplicationTemplate(ctx ...FieldCtx) string {
	return fmt.Sprintf(`
		id
		name
		description
		applicationInput
		placeholders {%s}
		webhooks {%s}
		labels
		accessLevel
	`, fp.ForPlaceholders(), fp.ForWebhooks())
}

// ForFormation missing godoc
func (fp *GqlFieldsProvider) ForFormation() string {
	return `
		id
		name
		formationTemplateId
	`
}

// ForFormationWithStatus missing godoc
func (fp *GqlFieldsProvider) ForFormationWithStatus() string {
	return fmt.Sprintf(`
		id
		name
		formationTemplateId
		status {%s}
	`, fp.ForFormationStatus())
}

// ForFormationTemplate missing godoc
func (fp *GqlFieldsProvider) ForFormationTemplate() string {
	return `
		id
		name
        applicationTypes
	    runtimeTypes
		runtimeTypeDisplayName	
		runtimeArtifactKind
	`
}

// ForFormationAssignment missing godoc
func (fp *GqlFieldsProvider) ForFormationAssignment() string {
	return `
			id
			source
			sourceType
			target
			targetType
			state
			value
	`
}

// ForFormationStatus missing godoc
func (fp *GqlFieldsProvider) ForFormationStatus() string {
	return fmt.Sprintf(`
			condition
			errors {
				%s
			}
	`, fp.ForFormationStatusErrors())
}

// ForFormationStatusErrors missing godoc
func (fp *GqlFieldsProvider) ForFormationStatusErrors() string {
	return `
        	assignmentID
        	message
        	errorCode
	`
}

// ForFormationConstraint missing godoc
func (fp *GqlFieldsProvider) ForFormationConstraint() string {
	return `
			id
			name
			constraintType
			targetOperation
			operator
			resourceType
			resourceSubtype
			inputTemplate
			constraintScope
	`
}

// ForFormationTemplateConstraintReference missing godoc
func (fp *GqlFieldsProvider) ForFormationTemplateConstraintReference() string {
	return `
			constraintID
			formationTemplateID
	`
}

// OmitForWebhooks missing godoc
func (fp *GqlFieldsProvider) OmitForWebhooks(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"id":               "id",
		"applicationID":    "applicationID",
		"type":             "type",
		"mode":             "mode",
		"correlationIdKey": "correlationIdKey",
		"retryInterval":    "retryInterval",
		"timeout":          "timeout",
		"url":              "url",
		"urlTemplate":      "urlTemplate",
		"inputTemplate":    "inputTemplate",
		"headerTemplate":   "headerTemplate",
		"outputTemplate":   "outputTemplate",
		"statusTemplate":   "statusTemplate",
		"auth":             fmt.Sprintf("auth {%s}", fp.ForAuth()),
		"createdAt":        "createdAt",
	}, omittedProperties)
}

// ForWebhooks missing godoc
func (fp *GqlFieldsProvider) ForWebhooks() string {
	return fmt.Sprintf(
		`id
		applicationID
		applicationTemplateID
		runtimeID
		type
		mode
		correlationIdKey
		retryInterval
		timeout
		url
		urlTemplate
		inputTemplate
		headerTemplate
		outputTemplate
		statusTemplate
		auth {
		  %s
		}
		createdAt`, fp.ForAuth())
}

// OmitForAPIDefinition missing godoc
func (fp *GqlFieldsProvider) OmitForAPIDefinition(omittedProperties []string) string {
	specOmittedProperties := extractOmitFor(omittedProperties, "spec")
	versionOmittedProperties := extractOmitFor(omittedProperties, "version")

	return buildProperties(map[string]string{
		"id":          "id",
		"name":        "name",
		"description": "description",
		"spec":        fmt.Sprintf("spec {%s}", fp.OmitForAPISpec(specOmittedProperties)),
		"targetURL":   "targetURL",
		"group":       "group",
		"version":     fmt.Sprintf("version {%s}", fp.OmitForVersion(versionOmittedProperties)),
	}, omittedProperties)
}

// ForAPIDefinition missing godoc
func (fp *GqlFieldsProvider) ForAPIDefinition(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		description
		spec {%s}
		targetURL
		group
		version {%s}`, fp.ForAPISpec(), fp.ForVersion()),
		ctx, []string{"APIDefinition.auth"})
}

// ForSystemAuth missing godoc
func (fp *GqlFieldsProvider) ForSystemAuth() string {
	return fmt.Sprintf(`
		id
		auth {%s}`, fp.ForAuth())
}

// OmitForAPISpec missing godoc
func (fp *GqlFieldsProvider) OmitForAPISpec(omittedProperties []string) string {
	frOmittedProperties := extractOmitFor(omittedProperties, "fetchRequest")

	return buildProperties(map[string]string{
		"id":           "id",
		"data":         "data",
		"format":       "format",
		"type":         "type",
		"fetchRequest": fmt.Sprintf("fetchRequest {%s}", fp.OmitForFetchRequest(frOmittedProperties)),
	}, omittedProperties)
}

// ForAPISpec missing godoc
func (fp *GqlFieldsProvider) ForAPISpec() string {
	return fmt.Sprintf(`
		id
		data
		format
		type
		fetchRequest {%s}`, fp.ForFetchRequest())
}

// OmitForFetchRequest missing godoc
func (fp *GqlFieldsProvider) OmitForFetchRequest(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"url":    "url",
		"auth":   fmt.Sprintf("auth {%s}", fp.ForAuth()),
		"mode":   "mode",
		"filter": "filter",
		"status": "status {condition message timestamp}",
	}, omittedProperties)
}

// ForFetchRequest missing godoc
func (fp *GqlFieldsProvider) ForFetchRequest() string {
	return fmt.Sprintf(`
		url
		auth {%s}
		mode
		filter
		status {condition message timestamp}`, fp.ForAuth())
}

// ForAPIRuntimeAuth missing godoc
func (fp *GqlFieldsProvider) ForAPIRuntimeAuth() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, fp.ForAuth())
}

// OmitForVersion missing godoc
func (fp *GqlFieldsProvider) OmitForVersion(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"value":           "value",
		"deprecated":      "deprecated",
		"deprecatedSince": "deprecatedSince",
		"forRemoval":      "forRemoval",
	}, omittedProperties)
}

// ForVersion missing godoc
func (fp *GqlFieldsProvider) ForVersion() string {
	return `value
		deprecated
		deprecatedSince
		forRemoval`
}

// ForPageInfo missing godoc
func (fp *GqlFieldsProvider) ForPageInfo() string {
	return `startCursor
		endCursor
		hasNextPage`
}

// OmitForEventDefinition missing godoc
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

// ForEventDefinition missing godoc
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

// OmitForEventSpec missing godoc
func (fp *GqlFieldsProvider) OmitForEventSpec(omittedProperties []string) string {
	frOmittedProperties := extractOmitFor(omittedProperties, "fetchRequest")

	return buildProperties(map[string]string{
		"id":           "id",
		"data":         "data",
		"type":         "type",
		"format":       "format",
		"fetchRequest": fmt.Sprintf("fetchRequest {%s}", fp.OmitForFetchRequest(frOmittedProperties)),
	}, omittedProperties)
}

// ForEventSpec missing godoc
func (fp *GqlFieldsProvider) ForEventSpec() string {
	return fmt.Sprintf(`
		id
		data
		type
		format
		fetchRequest {%s}`, fp.ForFetchRequest())
}

// OmitForDocument missing godoc
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

// ForDocument missing godoc
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

// ForAuth missing godoc
func (fp *GqlFieldsProvider) ForAuth() string {
	return `credential {
				... on BasicCredentialData {
					username
					password
				}
				...  on CertificateOAuthCredentialData {
					clientId
					certificate
					url
				}
   				...  on OAuthCredentialData {
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
				  ...  on OAuthCredentialData {
					clientId
					clientSecret
					url
				  }
				  ...  on CertificateOAuthCredentialData {
					clientId
					certificate
					url
				  }
			    }
				additionalHeaders
				additionalQueryParams
			  }
			}
		`
}

// ForLabel missing godoc
func (fp *GqlFieldsProvider) ForLabel() string {
	return `key
			value`
}

// ForRuntime missing godoc
func (fp *GqlFieldsProvider) ForRuntime(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		description
		labels 
		status {condition timestamp}
		metadata { creationTimestamp }
		auths {%s}
		runtimeContexts {%s}
        webhooks {%s}
		eventingConfiguration { defaultURL }`, fp.ForSystemAuth(), fp.Page(fp.ForRuntimeContext()), fp.ForWebhooks()), ctx, []string{"Runtime.runtimeContext"})
}

// ForRuntimeContext missing godoc
func (fp *GqlFieldsProvider) ForRuntimeContext() string {
	return `id
			key
			value
			labels`
}

// ForApplicationLabel missing godoc
func (fp *GqlFieldsProvider) ForApplicationLabel() string {
	return `
		key
		value`
}

// ForLabelDefinition missing godoc
func (fp *GqlFieldsProvider) ForLabelDefinition() string {
	return `
		key
		schema`
}

// ForOneTimeTokenForApplication missing godoc
func (fp *GqlFieldsProvider) ForOneTimeTokenForApplication() string {
	return `
		token
		connectorURL
		raw
		rawEncoded
		legacyConnectorURL`
}

// ForOneTimeTokenForRuntime missing godoc
func (fp *GqlFieldsProvider) ForOneTimeTokenForRuntime() string {
	return `
		token
		connectorURL
		raw
		rawEncoded`
}

// ForIntegrationSystem missing godoc
func (fp *GqlFieldsProvider) ForIntegrationSystem() string {
	return fmt.Sprintf(`
		id
		name
		description
		auths {%s}`, fp.ForSystemAuth())
}

// ForPlaceholders missing godoc
func (fp *GqlFieldsProvider) ForPlaceholders() string {
	return `
		name
		description
		jsonPath`
}

// ForEventingConfiguration missing godoc
func (fp *GqlFieldsProvider) ForEventingConfiguration() string {
	return `
		defaultURL`
}

// ForViewer missing godoc
func (fp *GqlFieldsProvider) ForViewer() string {
	return `
		id
		type`
}

// OmitForTenant missing godoc
func (fp *GqlFieldsProvider) OmitForTenant(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"id":          "id",
		"internalID":  "internalID",
		"name":        "name",
		"initialized": "initialized",
		"labels":      fmt.Sprintf("labels {%s}", fp.ForLabel()),
	}, omittedProperties)
}

// ForTenant missing godoc
func (fp *GqlFieldsProvider) ForTenant() string {
	return `
		id
		internalID
		name
		initialized
		labels
		parentID
		type`
}

// OmitForBundle missing godoc
func (fp *GqlFieldsProvider) OmitForBundle(omittedProperties []string) string {
	apiDefOmittedProperties := extractOmitFor(omittedProperties, "apiDefinitions")
	eventDefOmittedProperties := extractOmitFor(omittedProperties, "eventDefinitions")
	documentsOmittedProperties := extractOmitFor(omittedProperties, "documents")
	instanceAuthsOmittedProperties := extractOmitFor(omittedProperties, "instanceAuths")

	return buildProperties(map[string]string{
		"id":                             "id",
		"name":                           "name",
		"description":                    "description",
		"instanceAuthRequestInputSchema": "instanceAuthRequestInputSchema",
		"instanceAuths":                  fmt.Sprintf("instanceAuths {%s}", fp.OmitForBundleInstanceAuth(instanceAuthsOmittedProperties)),
		"defaultInstanceAuth":            fmt.Sprintf("defaultInstanceAuth {%s}", fp.ForAuth()),
		"apiDefinitions":                 fmt.Sprintf("apiDefinitions {%s}", fp.Page(fp.OmitForAPIDefinition(apiDefOmittedProperties))),
		"eventDefinitions":               fmt.Sprintf("eventDefinitions {%s}", fp.Page(fp.OmitForEventDefinition(eventDefOmittedProperties))),
		"documents":                      fmt.Sprintf("documents {%s}", fp.Page(fp.OmitForDocument(documentsOmittedProperties))),
	}, omittedProperties)
}

// ForBundle missing godoc
func (fp *GqlFieldsProvider) ForBundle(ctx ...FieldCtx) string {
	return addFieldsFromContext(fmt.Sprintf(`
		id
		name
		description
		instanceAuthRequestInputSchema
		instanceAuths {%s}
		defaultInstanceAuth {%s}
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}`, fp.ForBundleInstanceAuth(), fp.ForAuth(), fp.Page(fp.ForAPIDefinition(ctx...)), fp.Page(fp.ForEventDefinition()), fp.Page(fp.ForDocument())),
		ctx, []string{"Bundle.instanceAuth"})
}

// OmitForBundleInstanceAuth missing godoc
func (fp *GqlFieldsProvider) OmitForBundleInstanceAuth(omittedProperties []string) string {
	statusOmittedProperties := extractOmitFor(omittedProperties, "status")

	return buildProperties(map[string]string{
		"id":          "id",
		"context":     "context",
		"inputParams": "inputParams",
		"auth":        fmt.Sprintf("auth {%s}", fp.ForAuth()),
		"status":      fmt.Sprintf("status {%s}", fp.OmitForBundleInstanceAuthStatus(statusOmittedProperties)),
	}, omittedProperties)
}

// ForBundleInstanceAuth missing godoc
func (fp *GqlFieldsProvider) ForBundleInstanceAuth() string {
	return fmt.Sprintf(`
		id
		context
		inputParams
		auth {%s}
		status {%s}
		runtimeID
		runtimeContextID`, fp.ForAuth(), fp.ForBundleInstanceAuthStatus())
}

// OmitForBundleInstanceAuthStatus missing godoc
func (fp *GqlFieldsProvider) OmitForBundleInstanceAuthStatus(omittedProperties []string) string {
	return buildProperties(map[string]string{
		"condition": "condition",
		"timestamp": "timestamp",
		"message":   "message",
		"reason":    "reason",
	}, omittedProperties)
}

// ForBundleInstanceAuthStatus missing godoc
func (fp *GqlFieldsProvider) ForBundleInstanceAuthStatus() string {
	return `
		condition
		timestamp
		message
		reason`
}

// ForAutomaticScenarioAssignment missing godoc
func (fp *GqlFieldsProvider) ForAutomaticScenarioAssignment() string {
	return fmt.Sprintf(`
		scenarioName
		selector {%s}`, fp.ForLabel())
}
