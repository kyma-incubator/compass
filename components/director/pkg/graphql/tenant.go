package graphql

// Tenant missing godoc
type Tenant struct {
	ID          string   `json:"id"`
	InternalID  string   `json:"internalID"`
	Name        *string  `json:"name"`
	Type        string   `json:"type"`
	Parents     []string `json:"parents"`
	Initialized *bool    `json:"initialized"`
	Labels      Labels   `json:"labels"`
	Provider    string   `json:"provider"`
}
