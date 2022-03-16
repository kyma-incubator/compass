package model

import "github.com/kyma-incubator/compass/components/director/pkg/auth"

// APIRuntimeAuth missing godoc
type APIRuntimeAuth struct {
	ID        *string
	TenantID  string
	RuntimeID string
	APIDefID  string
	Value     *auth.Auth
}
