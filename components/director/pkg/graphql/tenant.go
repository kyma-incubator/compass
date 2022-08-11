package graphql

// Tenant missing godoc
type Tenant struct {
	ID          string  `json:"id"`
	InternalID  string  `json:"internalID"`
	Name        *string `json:"name"`
	Type        string  `json:"type"`
	ParentID    string  `json:"parentID"`
	Initialized *bool   `json:"initialized"`
	Labels      Labels  `json:"labels"`
	Provider    string  `json:"provider"`
}
