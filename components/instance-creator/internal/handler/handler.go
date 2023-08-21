package handler

import (
	"net/http"
)

// InstanceCreatorHandler processes received requests
type InstanceCreatorHandler struct{}

// HandlerFunc is the implementation of AdapterHandler
func (i InstanceCreatorHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
}
