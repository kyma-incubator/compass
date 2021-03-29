package fixtures

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

const (
	webhookURL          = "https://kyma-project.io"
	integrationSystemID = "69230297-3c81-4711-aac2-3afa8cb42e2d"
)

func FixSampleApplicationRegisterInputWithName(placeholder, name string) graphql.ApplicationRegisterInput {
	sampleInput := FixSampleApplicationRegisterInput(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func FixSampleApplicationRegisterInput(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: ptr.String("compass"),
		Labels:       &graphql.Labels{placeholder: []interface{}{placeholder}},
	}
}

func FixSampleApplicationRegisterInputWithWebhooks(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.WebhookTypeConfigurationChanged,
			URL:  ptr.String(webhookURL),
		},
		},
	}
}

func FixSampleApplicationRegisterInputWithNameAndWebhooks(placeholder, name string) graphql.ApplicationRegisterInput {
	sampleInput := FixSampleApplicationRegisterInputWithWebhooks(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func FixSampleApplicationCreateInputWithIntegrationSystem(placeholder string) graphql.ApplicationRegisterInput {
	sampleInput := FixSampleApplicationRegisterInputWithWebhooks(placeholder)
	sampleInput.IntegrationSystemID = ptr.String(integrationSystemID)
	return sampleInput
}

func FixSampleApplicationUpdateInput(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:    &placeholder,
		HealthCheckURL: ptr.String(webhookURL),
		ProviderName:   &placeholder,
	}
}

func FixSampleApplicationUpdateInputWithIntegrationSystem(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:         &placeholder,
		HealthCheckURL:      ptr.String(webhookURL),
		IntegrationSystemID: ptr.String(integrationSystemID),
		ProviderName:        ptr.String(placeholder),
	}
}

func FixApplicationRegisterInputWithBundles(t *testing.T) graphql.ApplicationRegisterInput {
	bndl1 := FixBundleCreateInputWithRelatedObjects(t, "foo")
	bndl2 := FixBundleCreateInputWithRelatedObjects(t, "bar")
	return graphql.ApplicationRegisterInput{
		Name:         "create-application-with-documents",
		ProviderName: ptr.String("compass"),
		Bundles: []*graphql.BundleCreateInput{
			&bndl1, &bndl2,
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
}

func FixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixGetApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixUpdateApplicationRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixUnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixAsyncUnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s" mode: ASYNC) {
			%s
		}	
	}`, id, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixRequestClientCredentialsForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestClientCredentialsForApplication(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixRequestOneTimeTokenForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForApplication(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForOneTimeTokenForApplication()))
}

func FixApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"") { 
					%s 
				}
			}`, runtimeID, 4, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixGetApplicationsRequestWithPagination() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications {
						%s
					}
				}`,
			testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixApplicationsFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixApplicationsPageableRequest(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixGetApplicationTemplatesWithPagination(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applicationTemplates(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplicationTemplate())))
}

func FixDeleteSystemAuthForApplicationRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForApplication(authID: "%s") {
					%s
				}
			}`, authID, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixRegisterApplicationFromTemplate(applicationFromTemplateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplicationFromTemplate(in: %s) {
					%s
				}
			}`,
			applicationFromTemplateInputInGQL, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixSetDefaultEventingForApplication(appID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setDefaultEventingForApplication(runtimeID: "%s", appID: "%s") {
					%s
				}
			}`,
			runtimeID, appID, testctx.Tc.GQLFieldsProvider.ForEventingConfiguration()))
}

func FixDeleteDefaultEventingForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteDefaultEventingForApplication(appID: "%s") {
						%s
					}
				}`,
			appID, testctx.Tc.GQLFieldsProvider.ForEventingConfiguration()))
}

// TODO: Delete after bundles are adopted
func FixRegisterApplicationWithPackagesRequest(name string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			  result: registerApplication(
				in: {
				  name: "%s"
				  providerName: "compass"
				  labels: { scenarios: ["DEFAULT"] }
				  packages: [
					{
					  name: "foo"
					  description: "Foo bar"
					  apiDefinitions: [
						{
						  name: "comments-v1"
						  description: "api for adding comments"
						  targetURL: "http://mywordpress.com/comments"
						  group: "comments"
						  spec: {
							data: "{\"openapi\":\"3.0.2\"}"
							type: OPEN_API
							format: YAML
						  }
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  targetURL: "http://mywordpress.com/reviews"
						  spec: {
							type: ODATA
							format: JSON
							fetchRequest: {
							  url: "http://mywordpress.com/apis"
							  auth: {
								credential: {
								  basic: { username: "admin", password: "secret" }
								}
								additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
								additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							  }
							  mode: PACKAGE
							  filter: "odata.json"
							}
						  }
						}
						{
						  name: "xml"
						  targetURL: "http://mywordpress.com/xml"
						  spec: { data: "odata", type: ODATA, format: XML }
						}
					  ]
					  eventDefinitions: [
						{
						  name: "comments-v1"
						  description: "comments events"
						  spec: {
							data: "{\"asyncapi\":\"1.2.0\"}"
							type: ASYNC_API
							format: YAML
						  }
						  group: "comments"
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  description: "review events"
						  spec: {
							type: ASYNC_API
							fetchRequest: {
							  url: "http://mywordpress.com/events"
							  auth: {
								credential: {
								  oauth: {
									clientId: "clientid"
									clientSecret: "grazynasecret"
									url: "url.net"
								  }
								}
							  }
							  mode: PACKAGE
							  filter: "async.json"
							}
							format: YAML
						  }
						}
					  ]
					  documents: [
						{
						  title: "Readme"
						  displayName: "display-name"
						  description: "Detailed description of project"
						  format: MARKDOWN
						  fetchRequest: {
							url: "kyma-project.io"
							auth: {
							  credential: {
								basic: { username: "admin", password: "secret" }
							  }
							  additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
							  additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							}
							mode: PACKAGE
							filter: "/docs/README.md"
						  }
						}
						{
						  title: "Troubleshooting"
						  displayName: "display-name"
						  description: "Troubleshooting description"
						  format: MARKDOWN
						  data: "No problems, everything works on my machine"
						}
					  ]
					}
					{
					  name: "bar"
					  description: "Foo bar"
					  apiDefinitions: [
						{
						  name: "comments-v1"
						  description: "api for adding comments"
						  targetURL: "http://mywordpress.com/comments"
						  group: "comments"
						  spec: {
							data: "{\"openapi\":\"3.0.2\"}"
							type: OPEN_API
							format: YAML
						  }
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  targetURL: "http://mywordpress.com/reviews"
						  spec: {
							type: ODATA
							format: JSON
							fetchRequest: {
							  url: "http://mywordpress.com/apis"
							  auth: {
								credential: {
								  basic: { username: "admin", password: "secret" }
								}
								additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
								additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							  }
							  mode: PACKAGE
							  filter: "odata.json"
							}
						  }
						}
						{
						  name: "xml"
						  targetURL: "http://mywordpress.com/xml"
						  spec: { data: "odata", type: ODATA, format: XML }
						}
					  ]
					  eventDefinitions: [
						{
						  name: "comments-v1"
						  description: "comments events"
						  spec: {
							data: "{\"asyncapi\":\"1.2.0\"}"
							type: ASYNC_API
							format: YAML
						  }
						  group: "comments"
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  description: "review events"
						  spec: {
							type: ASYNC_API
							fetchRequest: {
							  url: "http://mywordpress.com/events"
							  auth: {
								credential: {
								  oauth: {
									clientId: "clientid"
									clientSecret: "grazynasecret"
									url: "url.net"
								  }
								}
							  }
							  mode: PACKAGE
							  filter: "async.json"
							}
							format: YAML
						  }
						}
					  ]
					  documents: [
						{
						  title: "Readme"
						  displayName: "display-name"
						  description: "Detailed description of project"
						  format: MARKDOWN
						  fetchRequest: {
							url: "kyma-project.io"
							auth: {
							  credential: {
								basic: { username: "admin", password: "secret" }
							  }
							  additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
							  additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							}
							mode: PACKAGE
							filter: "/docs/README.md"
						  }
						}
						{
						  title: "Troubleshooting"
						  displayName: "display-name"
						  description: "Troubleshooting description"
						  format: MARKDOWN
						  data: "No problems, everything works on my machine"
						}
					  ]
					}
				  ]
				}
			  ) {
				id
				name
				providerName
				description
				integrationSystemID
				labels
				status {
				  condition
				  timestamp
				}
				webhooks {
				  id
				  applicationID
				  type
				  url
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
				healthCheckURL
				packages {
				  data {
					id
					name
					description
					instanceAuthRequestInputSchema
					instanceAuths {
					  id
					  context
					  inputParams
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
					  status {
						condition
						timestamp
						message
						reason
					  }
					}
					defaultInstanceAuth {
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
					apiDefinitions {
					  data {
						id
						name
						description
						spec {
						  data
						  format
						  type
						  fetchRequest {
							url
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
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						targetURL
						group
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					eventDefinitions {
					  data {
						id
						name
						description
						group
						spec {
						  data
						  type
						  format
						  fetchRequest {
							url
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
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					documents {
					  data {
						id
						title
						displayName
						description
						format
						kind
						data
						fetchRequest {
						  url
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
						  mode
						  filter
						  status {
							condition
							message
							timestamp
						  }
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
				  }
				  pageInfo {
					startCursor
					endCursor
					hasNextPage
				  }
				  totalCount
				}
				auths {
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
				eventingConfiguration {
				  defaultURL
				}
			  }
			}
		`, name))
}

// TODO: Delete after bundles are adopted
func FixGetApplicationWithPackageRequest(appID, packageID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			  result: application(id: "%s") {
				id
				name
				providerName
				description
				integrationSystemID
				labels
				status {
				  condition
				  timestamp
				}
				webhooks {
				  id
				  applicationID
				  type
				  url
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
				healthCheckURL
				package(id: "%s") {
					id
					name
					description
					instanceAuthRequestInputSchema
					instanceAuths {
					  id
					  context
					  inputParams
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
					  status {
						condition
						timestamp
						message
						reason
					  }
					}
					defaultInstanceAuth {
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
					apiDefinitions {
					  data {
						id
						name
						description
						spec {
						  data
						  format
						  type
						  fetchRequest {
							url
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
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						targetURL
						group
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					eventDefinitions {
					  data {
						id
						name
						description
						group
						spec {
						  data
						  type
						  format
						  fetchRequest {
							url
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
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					documents {
					  data {
						id
						title
						displayName
						description
						format
						kind
						data
						fetchRequest {
						  url
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
						  mode
						  filter
						  status {
							condition
							message
							timestamp
						  }
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
				}
				auths {
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
				eventingConfiguration {
				  defaultURL
				}
			  }
			}`, appID, packageID))
}

// TODO: Delete after bundles are adopted
func FixApplicationsForRuntimeWithPackagesRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: applicationsForRuntime(runtimeID: "%s") {
					data {
					  id
					  name
					  providerName
					  description
					  integrationSystemID
					  labels
					  status {
						condition
						timestamp
					  }
					  webhooks {
						id
						applicationID
						type
						url
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
					  healthCheckURL
					  packages {
						data {
						  id
						  name
						  description
						  instanceAuthRequestInputSchema
						  instanceAuths {
							id
							context
							inputParams
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
							status {
							  condition
							  timestamp
							  message
							  reason
							}
						  }
						  defaultInstanceAuth {
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
						  apiDefinitions {
							data {
							  id
							  name
							  description
							  spec {
								data
								format
								type
								fetchRequest {
								  url
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
								  mode
								  filter
								  status {
									condition
									message
									timestamp
								  }
								}
							  }
							  targetURL
							  group
							  version {
								value
								deprecated
								deprecatedSince
								forRemoval
							  }
							}
							pageInfo {
							  startCursor
							  endCursor
							  hasNextPage
							}
							totalCount
						  }
						  eventDefinitions {
							data {
							  id
							  name
							  description
							  group
							  spec {
								data
								type
								format
								fetchRequest {
								  url
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
								  mode
								  filter
								  status {
									condition
									message
									timestamp
								  }
								}
							  }
							  version {
								value
								deprecated
								deprecatedSince
								forRemoval
							  }
							}
							pageInfo {
							  startCursor
							  endCursor
							  hasNextPage
							}
							totalCount
						  }
						  documents {
							data {
							  id
							  title
							  displayName
							  description
							  format
							  kind
							  data
							  fetchRequest {
								url
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
								mode
								filter
								status {
								  condition
								  message
								  timestamp
								}
							  }
							}
							pageInfo {
							  startCursor
							  endCursor
							  hasNextPage
							}
							totalCount
						  }
						}
						pageInfo {
						  startCursor
						  endCursor
						  hasNextPage
						}
						totalCount
					  }
					  auths {
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
					  eventingConfiguration {
						defaultURL
					  }
					}
					pageInfo {
					  startCursor
					  endCursor
					  hasNextPage
					}
					totalCount
				  }
				}`, runtimeID))
}
