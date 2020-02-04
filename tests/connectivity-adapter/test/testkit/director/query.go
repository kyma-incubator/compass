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

func setDefaultEventingMutation(runtimeID string, appID string) string {
	return fmt.Sprintf(`setDefaultEventingForApplication(appID: $appID, runtimeID: $runtime2) {
		defaultURL
	}
}`)
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
