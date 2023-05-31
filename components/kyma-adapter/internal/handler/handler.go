package handler

import (
	"net/http"

	"github.com/machinebox/graphql"
)

// AdapterHandler processes received requests
type AdapterHandler struct {
	GqlClient *graphql.Client
}

// HandlerFunc is the implementation of AdapterHandler
func (a AdapterHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
}
