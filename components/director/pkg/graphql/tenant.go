package graphql

// Tenant missing godoc
type Tenant struct {
	ID          string  `json:"id"`
	InternalID  string  `json:"internalID"`
	Name        *string `json:"name"`
	Initialized *bool   `json:"initialized"`
	Labels      Labels  `json:"labels"`
}
