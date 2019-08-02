package graphql

type Runtime struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Status      *RuntimeStatus `json:"status"`
	// TODO: directive for checking auth
	AgentAuth *Auth `json:"agentAuth"`
}
