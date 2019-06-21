package end_to_end

import "fmt"

type fieldsProvider struct{}

func (fp *fieldsProvider) Page(item string) string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, item, fp.ForPageInfo())
}

func (fp *fieldsProvider) ForApplication() string {
	return fmt.Sprintf(`id
		tenant
		name
		description
		labels
		annotations
		status {condition timestamp}
		webhooks {%s}
		healthCheckURL
		apis {%s}
		eventAPIs {%s}
		documents {%s}
	`, fp.ForWebhooks(), fp.ForApiDefinitionPage(), fp.ForEventAPIPage(), fp.ForDocumentPage())
}

func (fp *fieldsProvider) ForWebhooks() string {
	return fmt.Sprintf(
		`id
		type
		url
		auth {
		  %s
		}`, fp.ForAuth())
}

func (fp *fieldsProvider) ForAPIDefinition() string {
	return fmt.Sprintf(`		id
		name
		description
		spec {%s}
		targetURL
		group
		auths {%s}
		defaultAuth {%s}
		version {%s}`, fp.ForApiSpec(), fp.ForAuthRuntime(), fp.ForAuth(), fp.ForVersion())
}

func (fp *fieldsProvider) ForApiDefinitionPage() string {
	return fmt.Sprintf(`data {
		%s
	}
	pageInfo {%s}
	totalCount
	`, fp.ForAPIDefinition(), fp.ForPageInfo())
}

func (fp *fieldsProvider) ForApiSpec() string {
	return fmt.Sprintf(`data
		format
		type
		fetchRequest {%s}`, fp.ForFetchRequest())
}

func (fp *fieldsProvider) ForFetchRequest() string {
	return fmt.Sprintf(`url
		auth {%s}
		mode
		filter
		status {condition timestamp}`, fp.ForAuth())
}

func (fp *fieldsProvider) ForAuthRuntime() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, fp.ForAuth())
}

func (fp *fieldsProvider) ForVersion() string {
	return `value
		deprecated
		deprecatedSince
		forRemoval`
}

func (fp *fieldsProvider) ForPageInfo() string {
	return `startCursor
		endCursor
		hasNextPage`
}

func (fp *fieldsProvider) ForEventAPIPage() string {
	return fmt.Sprintf(`data {
			id
			name
			description
			group 
			spec {%s}
			version {%s}
		}
		pageInfo {%s}
		totalCount`, fp.ForEventSpec(), fp.ForVersion(), fp.ForPageInfo())
}

func (fp *fieldsProvider) ForEventSpec() string {
	return fmt.Sprintf(`data
		type
		format
		fetchRequest {%s}`, fp.ForFetchRequest())
}
func (fp *fieldsProvider) ForDocumentPage() string {
	return fmt.Sprintf(`data {
		id
		title
		format
		kind
		data
		fetchRequest {%s}
	}
	pageInfo {%s}
	totalCount
	`, fp.ForFetchRequest(), fp.ForPageInfo())
}

func (fp *fieldsProvider) ForAuth() string {
	// TODO request auth
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
		`, )
}

func (fp *fieldsProvider) ForRuntime() string {
	return fmt.Sprintf(`id
		name
		description
		tenant
		labels 
		annotations
		status {condition timestamp}`)
	//TODO agentAuth {%s}`, fp.ForAuth())
}
