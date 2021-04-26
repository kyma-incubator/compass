package graphql

type Runtime struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Description           *string                       `json:"description"`
	Status                *RuntimeStatus                `json:"status"`
	Metadata              *RuntimeMetadata              `json:"metadata"`
	EventingConfiguration *RuntimeEventingConfiguration `json:"eventingConfiguration"`
}

// Extended types used by external API

type RuntimePageExt struct {
	RuntimePage
	Data []*RuntimeExt `json:"data"`
}

type RuntimeExt struct {
	Runtime
	Labels Labels `json:"labels"`
	// Returns array of authentication details for Runtime. For now at most one element in array will be returned.
	Auths []*RuntimeSystemAuth `json:"auths"`
}
