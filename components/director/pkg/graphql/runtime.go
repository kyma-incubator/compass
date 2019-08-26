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
	Data       []*RuntimeExt     `json:"data"`
	PageInfo   *PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type RuntimeExt struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Status      *RuntimeStatus `json:"status"`
	// TODO: directive for checking auth
	AgentAuth *Auth `json:"agentAuth"`
	Labels Labels `json:"labels"`
}
