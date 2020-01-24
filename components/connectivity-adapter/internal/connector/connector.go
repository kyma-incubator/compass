package connector

import (
	"github.com/gorilla/mux"
)

type Config struct{}

func RegisterHandler(router *mux.Router, cfg Config) {
	// TODO: Define endpoints for Connector Service adapter
}
