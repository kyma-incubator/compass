package handler

import (
	"net/http"
)

//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	// todo::: add SM calls here
}

// InstanceCreatorHandler processes received requests
type InstanceCreatorHandler struct {
	SMClient Client
}

// NewHandler creates an InstanceCreatorHandler
func NewHandler(smClient Client) *InstanceCreatorHandler {
	return &InstanceCreatorHandler{
		SMClient: smClient,
	}
}

// HandlerFunc is the implementation of AdapterHandler
func (i InstanceCreatorHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	// todo::: note: here you will call i.SMClient.Method()
}
