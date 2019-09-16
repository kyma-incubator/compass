package graphql

type Runtime struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Status      *RuntimeStatus `json:"status"`
	// TODO: directive for checking auth
	AgentAuth *Auth `json:"agentAuth"`
}

// Extended types used by external API

type RuntimePageExt struct {
	RuntimePage
	Data []*RuntimeExt `json:"data"`
}

type RuntimeExt struct {
	Runtime
	Labels Labels `json:"labels"`
}
