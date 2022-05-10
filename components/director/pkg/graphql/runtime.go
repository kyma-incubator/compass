package graphql

// Runtime missing godoc
type Runtime struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Description           *string                       `json:"description"`
	Status                *RuntimeStatus                `json:"status"`
	Metadata              *RuntimeMetadata              `json:"metadata"`
	EventingConfiguration *RuntimeEventingConfiguration `json:"eventingConfiguration"`
}

// RuntimePageExt is an extended types used by external API
type RuntimePageExt struct {
	RuntimePage
	Data []*RuntimeExt `json:"data"`
}

// RuntimeExt missing godoc
type RuntimeExt struct {
	Runtime
	Labels Labels `json:"labels"`
	// Returns array of authentication details for Runtime. For now at most one element in array will be returned.
	Auths           []*RuntimeSystemAuth  `json:"auths"`
	RuntimeContext  RuntimeContextExt     `json:"runtimeContext"`
	RuntimeContexts RuntimeContextPageExt `json:"runtimeContexts"`
}
