package model

type RuntimeAuth struct {
	ID        *string
	TenantID  string
	RuntimeID string
	APIDefID  string
	Value     *Auth
}
