package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) registerRuntime(input string) string {
	return fmt.Sprintf(`mutation {
	result: createRuntime(in: %s) {
		%s
	}
}`, input, runtimeData())
}

func (qp queryProvider) deleteRuntime(runtimeId string) string {
	return fmt.Sprintf(`mutation {
  result: deleteRuntime(id: "%s") { id }
}`, runtimeId)
}

func runtimeData() string {
	return fmt.Sprintf(`
		id
		name
		description
		labels 
		status {condition timestamp}
		auths {%s}`, systemAuthData())
}

func systemAuthData() string {
	return fmt.Sprintf(`
		id
		auth {%s}`, authData())
}

func authData() string {
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

func apiSpecData() string {
	return `data
		format
		type`
}

func runtimeAuthData() string {
	return fmt.Sprintf(`runtimeID
		auth {%s}`, authData())
}
