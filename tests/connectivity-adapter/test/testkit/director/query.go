package director

import "fmt"

func createApplicationQuery(in string) string {
	return fmt.Sprintf(`mutation{
 		result: registerApplication(in: %s)
        {
			id
			name
			description
			labels
			eventingConfiguration { defaultURL } 
		}
   }`, in)
}

func createRuntimeQuery(in string) string {
	return fmt.Sprintf(`mutation{
    result: registerRuntime(in: %s)
	{
		id
	}
	}`, in)
}

func setApplicationLabel(applicationID, key, value string) string {
	return fmt.Sprintf(`mutation {
  	result: setApplicationLabel(applicationID: "%s", key: "%s", value: "%s") {
    key
    value
  }
}`, applicationID, key, value)
}

func deleteApplicationLabel(applicationID, key string) string {
	return fmt.Sprintf(`mutation {
  	result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
    key
    value
  }
}`, applicationID, key)
}

func setRuntimeLabel(runtimeID, key, value string) string {
	return fmt.Sprintf(`mutation {
  	result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: "%s") {
    key
    value
  }
}`, runtimeID, key, value)
}

func setDefaultEventingQuery(runtimeID, appID string) string {
	return fmt.Sprintf(`mutation {
  	result: setDefaultEventingForApplication(appID: "%s",runtimeID: "%s") {
     defaultURL
  }
}`, appID, runtimeID)
}

func getOneTimeTokenQuery(appID string) string {
	return fmt.Sprintf(`mutation {
  	result: requestOneTimeTokenForApplication(id: "%s") {
      token
      connectorURL
	  legacyConnectorURL
  }
}`, appID)
}

func deleteApplicationQuery(appID string) string {
	return fmt.Sprintf(`mutation{
 		result: unregisterApplication(id: "%s")
        {
			id
		}
   }`, appID)
}

func deleteRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`mutation{
 		result: unregisterRuntime(id: "%s")
        {
			id
		}
   }`, runtimeID)
}

func getTenantsQuery() string {
	return `query {
		result: tenants {
			id
			name
			internalID
		}
	}`
}
