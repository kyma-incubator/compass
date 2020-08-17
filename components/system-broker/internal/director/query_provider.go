/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package director

import (
	"fmt"
)

//TODO first attempt at querying, probably outputGraphqlizer can do the same and this file can be removed
type queryProvider struct{}

func (qp queryProvider) applicationsForRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
	result: applicationsForRuntime(runtimeID: "%s") {
		%s
	}
}`, runtimeID, applicationsQueryData(runtimeID))
}

func applicationsQueryData(runtimeID string) string {
	return pageData(applicationData(runtimeID))
}

func pageData(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, pageInfoData())
}

func pageInfoData() string {
	return `startCursor
		endCursor
		hasNextPage`
}

func applicationData(runtimeID string) string {
	return fmt.Sprintf(`id
		name
		providerName
		description
		labels
		auths {%s}
		packages {%s}
	`, systemAuthData(), pageData(packagesData()))
}

func systemAuthData() string {
	return fmt.Sprintf(`id`)
}

func packagesData() string {
	return fmt.Sprintf(`id
		name
		description
		instanceAuthRequestInputSchema
		apiDefinitions {%s}
		eventDefinitions {%s}
		documents {%s}
		`, pageData(packageApiDefinitions()), pageData(eventAPIData()), pageData(documentData()))
}

func packageApiDefinitions() string {
	return fmt.Sprintf(`		id
		name
		description
		spec {%s}
		targetURL
		group
		version {%s}`, apiSpecData(), versionData())
}

func apiSpecData() string {
	return fmt.Sprintf(`data
		format
		type`)
}

func versionData() string {
	return `value
		deprecated
		deprecatedSince
		forRemoval`
}

func eventAPIData() string {
	return fmt.Sprintf(`
			id
			name
			description
			group 
			spec {%s}
			version {%s}
		`, eventSpecData(), versionData())
}

func eventSpecData() string {
	return fmt.Sprintf(`data
		type
		format`)
}

func documentData() string {
	return fmt.Sprintf(`
		id
		title
		displayName
		description
		format
		kind
		data`)
}

func (qp queryProvider) createRuntimeMutation(runtimeInput string) string {
	return fmt.Sprintf(`mutation {
	result: registerRuntime(in: %s) { id } }`, runtimeInput)
}

func (qp queryProvider) getRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
    result: runtime(id: "%s") {
         id name description labels
}}`, runtimeID)
}

func (qp queryProvider) deleteRuntimeMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterRuntime(id: "%s") {
		id
}}`, runtimeID)
}

func (qp queryProvider) requestOneTimeTokeneMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: requestOneTimeTokenForRuntime(id: "%s") {
		token connectorURL
}}`, runtimeID)
}
