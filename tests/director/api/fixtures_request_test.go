package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	gcli "github.com/machinebox/graphql"
)

func fixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.gqlFieldsProvider.ForApplication()))
}

// TODO: Delete after bundles are adopted
func fixRegisterApplicationWithPackagesRequest(name string) *gcli.Request {
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
func fixGetApplicationWithPackageRequest(appID, packageID string) *gcli.Request {
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
func fixApplicationsForRuntimeWithPackagesRequest(runtimeID string) *gcli.Request {
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

func fixUnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			unregisterApplication(id: "%s") {
				id
			}
		}`, id))
}

func fixCreateApplicationTemplateRequest(applicationTemplateInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplicationTemplate(in: %s) {
					%s
				}
			}`,
			applicationTemplateInGQL, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixRegisterRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerRuntime(in: %s) {
					%s
				}
			}`,
			runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
}

func fixCreateLabelDefinitionRequest(labelDefinitionInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createLabelDefinition(in: %s) {
						%s
					}
				}`,
			labelDefinitionInputGQL, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixRegisterIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixAddDocumentRequest(appID, documentInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addDocument(applicationID: "%s", in: %s) {
 				%s
			}				
		}`, appID, documentInputInGQL, tc.gqlFieldsProvider.ForDocument()))
}

func fixDeleteDocumentRequest(docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteDocument(id: "%s") {
					id
				}
			}`, docID))
}

func fixAddDocumentToBundleRequest(bundleID, documentInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addDocumentToBundle(bundleID: "%s", in: %s) {
 				%s
			}				
		}`, bundleID, documentInputInGQL, tc.gqlFieldsProvider.ForDocument()))
}

func fixAddWebhookRequest(applicationID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`,
			applicationID, webhookInGQL, tc.gqlFieldsProvider.ForWebhooks()))
}

func fixDeleteWebhookRequest(webhookID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteWebhook(webhookID: "%s") {
				%s
			}
		}`, webhookID, tc.gqlFieldsProvider.ForWebhooks()))
}

func fixAddAPIRequest(appID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinition(applicationID: "%s", in: %s) {
				%s
			}
		}
		`, appID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixAddAPIToBundleRequest(bndlID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bndlID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixUpdateAPIRequest(apiID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateAPIDefinition(id: "%s", in: %s) {
				%s
			}
		}
		`, apiID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixDeleteAPIRequest(apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: deleteAPIDefinition(id: "%s") {
				id
			}
		}`, apiID))
}

func fixAddEventAPIRequest(appID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventDefinition(applicationID: "%s", in: %s) {
				%s
			}
		}
		`, appID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixAddEventAPIToBundleRequest(bndlID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bndlID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixUpdateEventAPIRequest(eventAPIID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateEventDefinition(id: "%s", in: %s) {
				%s
			}
		}
		`, eventAPIID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixUpdateDocumentRequest(documentID, documentInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateDocument(id: "%s", in: %s) {
				%s
			}
		}
		`, documentID, documentInputGQL, tc.gqlFieldsProvider.ForDocument()))
}

func fixDeleteEventAPIRequest(eventAPIID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteEventDefinition(id: "%s") {
				id
			}
		}`, eventAPIID))
}

func fixUpdateRuntimeRequest(id, updateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateRuntime(id: "%s", in: %s) {
					%s
				}
			}`,
			id, updateInputInGQL, tc.gqlFieldsProvider.ForRuntime()))
}

func fixUpdateWebhookRequest(webhookID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`,
			webhookID, webhookInGQL, tc.gqlFieldsProvider.ForWebhooks()))
}

func fixUpdateApplicationRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixUpdateApplicationTemplateRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplicationTemplate(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixUpdateLabelDefinitionRequest(ldInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: updateLabelDefinition(in: %s) {
						%s
					}
				}`, ldInputGQL, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixRequestOneTimeTokenForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestOneTimeTokenForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForOneTimeTokenForRuntime()))
}

func fixRequestOneTimeTokenForApp(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestOneTimeTokenForApplication(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForOneTimeTokenForApplication()))
}

func fixUpdateIntegrationSystemRequest(id, integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateIntegrationSystem(id: "%s", in: %s) {
    					%s
					}
				}`, id, integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixRequestClientCredentialsForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForApplication(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixRequestClientCredentialsForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixRequestClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixSetApplicationLabelRequest(appID, labelKey string, labelValue interface{}) *gcli.Request {
	jsonValue, err := json.Marshal(labelValue)
	if err != nil {
		panic(errors.New("label value can not be marshalled"))
	}
	value := removeDoubleQuotesFromJSONKeys(string(jsonValue))

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
						%s
					}
				}`,
			appID, labelKey, value, tc.gqlFieldsProvider.ForLabel()))
}

func fixSetRuntimeLabelRequest(runtimeID, labelKey string, labelValue interface{}) *gcli.Request {
	jsonValue, err := json.Marshal(labelValue)
	if err != nil {
		panic(errors.New("label value can not be marshalled"))
	}
	value := removeDoubleQuotesFromJSONKeys(string(jsonValue))

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: %s) {
						%s
					}
				}`, runtimeID, labelKey, value, tc.gqlFieldsProvider.ForLabel()))
}

func fixSetAPIAuthRequest(apiID string, rtmID string, authInStr string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setAPIAuth(apiID: "%s", runtimeID: "%s", in: %s) {
					%s
				}
			}`, apiID, rtmID, authInStr, tc.gqlFieldsProvider.ForAPIRuntimeAuth()))
}

func fixApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"") { 
					%s 
				}
			}`, runtimeID, 4, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixRuntimeRequestWithPaginationRequest(after int, cursor string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes(first:%d, after:"%s") {
					%s
				}
			}`, after, cursor, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixAPIRuntimeAuthRequest(applicationID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
			}
		}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"APIDefinition.auth": fmt.Sprintf(`auth(runtimeID: "%s") {%s}`, runtimeID, tc.gqlFieldsProvider.ForAPIRuntimeAuth()),
		})))
}

func fixRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, runtimeID, tc.gqlFieldsProvider.ForRuntime()))
}

func fixUnregisterRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{unregisterRuntime(id: "%s") {
				id
			}
		}`, id))
}

func fixLabelDefinitionRequest(labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: labelDefinition(key: "%s") {
						%s
					}
				}`,
			labelKey, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixApplicationRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication()))
}

func fixApplicationTemplateRequest(applicationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applicationTemplate(id: "%s") {
					%s
				}
			}`, applicationTemplateID, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixLabelDefinitionsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result:	labelDefinitions() {
					key
					schema
				}
			}`))
}

func fixApplicationsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications {
						%s
					}
				}`,
			tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixApplicationsFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixApplicationsPageableRequest(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixApplicationTemplates(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applicationTemplates(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplicationTemplate())))
}

func fixRuntimesRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes {
						%s
					}
				}`,
			tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixRuntimesFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixIntegrationSystemsRequest(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: integrationSystems(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForIntegrationSystem())))
}

func fixIntegrationSystemRequest(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: integrationSystem(id: "%s") {
						%s
					}
				}`,
			intSysID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixTenantsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants {
						%s
					}
				}`, tc.gqlFieldsProvider.ForTenant()))
}

func fixDeleteLabelDefinitionRequest(labelDefinitionKey string, deleteRelatedLabels bool) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteLabelDefinition(key: "%s", deleteRelatedLabels: %t) {
					%s
				}
			}`, labelDefinitionKey, deleteRelatedLabels, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixDeleteRuntimeLabelRequest(runtimeID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
					%s
				}
			}`, runtimeID, labelKey, tc.gqlFieldsProvider.ForLabel()))
}

func fixDeleteApplicationLabelRequest(applicationID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
					%s
				}
			}`, applicationID, labelKey, tc.gqlFieldsProvider.ForLabel()))
}

func fixDeleteAPIAuthRequestRequest(apiID string, rtmID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteAPIAuth(apiID: "%s",runtimeID: "%s") {
					%s
				} 
			}`, apiID, rtmID, tc.gqlFieldsProvider.ForAPIRuntimeAuth()))
}

func fixunregisterIntegrationSystem(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterIntegrationSystem(id: "%s") {
					%s
				}
			}`, intSysID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixDeleteSystemAuthForApplicationRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForApplication(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixDeleteSystemAuthForRuntimeRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForRuntime(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixDeleteSystemAuthForIntegrationSystemRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForIntegrationSystem(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixDeleteApplicationTemplate(appTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationTemplate(id: "%s") {
					%s
				}
			}`, appTemplateID, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func removeDoubleQuotesFromJSONKeys(in string) string {
	var validRegex = regexp.MustCompile(`"(\w+|\$\w+)"\s*:`)
	return validRegex.ReplaceAllString(in, `$1:`)
}

func fixRegisterApplicationFromTemplate(applicationFromTemplateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplicationFromTemplate(in: %s) {
					%s
				}
			}`,
			applicationFromTemplateInputInGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixSetDefaultEventingForApplication(appID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setDefaultEventingForApplication(runtimeID: "%s", appID: "%s") {
					%s
				}
			}`,
			runtimeID, appID, tc.gqlFieldsProvider.ForEventingConfiguration()))
}

func fixAPIDefinitionInBundleRequest(appID, bndlID, apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							apiDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, apiID, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixEventDefinitionInBundleRequest(appID, bndlID, eventID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							eventDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, eventID, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixDocumentInBundleRequest(appID, bndlID, docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							document(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, docID, tc.gqlFieldsProvider.ForDocument()))
}

func fixAPIDefinitionsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							apiDefinitions{
						%s
						}					
					}
				}
			}`, appID, bndlID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForAPIDefinition())))

}

func fixEventDefinitionsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							eventDefinitions{
						%s
						}					
					}
				}
			}`, appID, bndlID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForEventDefinition())))
}

func fixDocumentsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							documents{
						%s
						}					
					}
				}
			}`, appID, bndlID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForDocument())))
}

func fixAddBundleRequest(appID, bndlCreateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addBundle(applicationID: "%s", in: %s) {
				%s
			}}`, appID, bndlCreateInput, tc.gqlFieldsProvider.ForBundle()))
}

func fixUpdateBundleRequest(bundleID, bndlUpdateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateBundle(id: "%s", in: %s) {
				%s
			}
		}`, bundleID, bndlUpdateInput, tc.gqlFieldsProvider.ForBundle()))
}

func fixDeleteBundleRequest(bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundle(id: "%s") {
				%s
			}
		}`, bundleID, tc.gqlFieldsProvider.ForBundle()))
}

func fixSetBundleInstanceAuthRequest(authID, apiAuthInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setBundleInstanceAuth(authID: "%s", in: %s) {
				%s
			}
		}`, authID, apiAuthInput, tc.gqlFieldsProvider.ForBundleInstanceAuth()))
}

func fixDeleteBundleInstanceAuthRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundleInstanceAuth(authID: "%s") {
				%s
			}
		}`, authID, tc.gqlFieldsProvider.ForBundleInstanceAuth()))
}

func fixRequestBundleInstanceAuthCreationRequest(bundleID, bndlInstanceAuthRequestInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthCreation(bundleID: "%s", in: %s) {
				%s
			}
		}`, bundleID, bndlInstanceAuthRequestInput, tc.gqlFieldsProvider.ForBundleInstanceAuth()))
}

func fixRequestBundleInstanceAuthDeletionRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthDeletion(authID: "%s") {
				%s
			}
		}`, authID, tc.gqlFieldsProvider.ForBundleInstanceAuth()))
}

func fixBundleRequest(applicationID string, bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`, bundleID, tc.gqlFieldsProvider.ForBundle()),
		})))
}

func fixAPIDefinitionRequest(applicationID string, apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.apiDefinition": fmt.Sprintf(`apiDefinition(id: "%s") {%s}`, apiID, tc.gqlFieldsProvider.ForAPIDefinition()),
		})))
}

func fixEventDefinitionRequest(applicationID string, eventDefID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.eventDefinition": fmt.Sprintf(`eventDefinition(id: "%s") {%s}`, eventDefID, tc.gqlFieldsProvider.ForEventDefinition()),
		})))
}

func fixBundleWithInstanceAuthRequest(applicationID string, bundleID string, instanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID,
			tc.gqlFieldsProvider.ForApplication(
				graphqlizer.FieldCtx{"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`,
					bundleID,
					tc.gqlFieldsProvider.ForBundle(graphqlizer.FieldCtx{
						"Bundle.instanceAuth": fmt.Sprintf(`instanceAuth(id: "%s") {%s}`,
							instanceAuthID,
							tc.gqlFieldsProvider.ForBundleInstanceAuth()),
					})),
				})))
}

func fixBundlesRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication()))
}

func fixDeleteDefaultEventingForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteDefaultEventingForApplication(appID: "%s") {
						%s
					}
				}`,
			appID, tc.gqlFieldsProvider.ForEventingConfiguration()))
}

func fixCreateAutomaticScenarioAssignmentRequest(automaticScenarioAssignmentInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createAutomaticScenarioAssignment(in: %s) {
						%s
					}
				}`,
			automaticScenarioAssignmentInput, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenario string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
            result: deleteAutomaticScenarioAssignmentForScenario(scenarioName: "%s") {
                  %s
               }
            }`,
			scenario, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixDeleteAutomaticScenarioAssignmentsForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
            result: deleteAutomaticScenarioAssignmentsForSelector(selector: %s) {
                  %s
               }
            }`,
			labelSelectorInput, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}
func fixAutomaticScenarioAssignmentsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignments {
						%s
					}
				}`,
			tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForAutomaticScenarioAssignment())))
}

func fixAutomaticScenarioAssignmentsForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentsForSelector(selector: %s) {
						%s
					}
				}`,
			labelSelectorInput, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixAutomaticScenarioAssignmentForScenarioRequest(scenarioName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentForScenario(scenarioName: "%s") {
						%s
					}
				}`,
			scenarioName, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixRefetchAPISpecRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: refetchAPISpec(apiID: "%s") {
						%s
					}
				}`,
			id, tc.gqlFieldsProvider.ForApiSpec()))
}
