package model

// APIRuntimeAuth missing godoc
type APIRuntimeAuth struct {
	ID        *string
	TenantID  string
	RuntimeID string
	APIDefID  string
	Value     *Auth
}
