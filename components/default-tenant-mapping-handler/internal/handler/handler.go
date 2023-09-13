package handler

import "net/http"

// DefaultTenantMappingHandler processes received requests
type DefaultTenantMappingHandler struct{}

// NewHandler creates an DefaultTenantMappingHandler
func NewHandler() *DefaultTenantMappingHandler {
	return &DefaultTenantMappingHandler{}
}

// HandlerFunc is the implementation of DefaultTenantMappingHandler
func (i DefaultTenantMappingHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {

}
