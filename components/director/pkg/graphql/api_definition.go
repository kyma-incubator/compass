package graphql

type APIDefinition struct {
	ID            string   `json:"id"`
	ApplicationID string   `json:"applicationID"`
	Name          string   `json:"name"`
	Description   *string  `json:"description"`
	Spec          *APISpec `json:"spec"`
	TargetURL     string   `json:"targetURL"`
	//  group allows you to find the same API but in different version
	Group *string `json:"group"`
	// If defaultAuth is specified, it will be used for all Runtimes that does not specify Auth explicitly.
	DefaultAuth *Auth    `json:"defaultAuth"`
	Version     *Version `json:"version"`
}
