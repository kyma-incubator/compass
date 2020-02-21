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

func setEventBaseURLQuery(runtimeID string, url string) string {
	return fmt.Sprintf(`mutation {
  	result: setRuntimeLabel(runtimeID: "%s", key: "runtime/event_service_url", value: "%s") {
    key
    value
  }
}`, runtimeID, url)
}

func setDefaultEventingQuery(runtimeID string, appID string) string {
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
