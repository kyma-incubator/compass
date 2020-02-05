package director

import "fmt"

func createApplicationMutation(in string) string {
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

func createRuntimeMutation(in string) string {
	return fmt.Sprintf(`mutation{
    result: registerRuntime(in: %s)
	{
		id
	}
	}`, in)
}

func setEventBaseURLMutation(runtimeID string, url string) string {
	return fmt.Sprintf(`mutation {
  	result: setRuntimeLabel(runtimeID: "%s", key: "runtime/event_service_url", value: "%s") {
    key
    value
  }
}`, runtimeID, url)
}

func setDefaultEventingForApplication(runtimeID string, appID string) string {
	return fmt.Sprintf(`mutation {
  	result: setDefaultEventingForApplication(appID: "%s",runtimeID: "%s") {
     defaultURL
  }
}`, appID, runtimeID)
}

func getOneTimeTokenForApplication(appID string) string {
	return fmt.Sprintf(`mutation {
  	result: requestOneTimeTokenForApplication(id: "%s") {
      token
      connectorURL
	  legacyConnectorURL
  }
}`, appID)
}

func applicationQuery(appID string) string {
	return fmt.Sprintf(`query{
 		result: application(id: "%s")
        {
 			 name
			 eventingConfiguration {
  				defaultURL
			}
		}
   }`, appID)
}
