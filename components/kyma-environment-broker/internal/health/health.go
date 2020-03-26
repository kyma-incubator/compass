package health

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HealthHandler struct {
	Address string
}

func NewHandler(addr string) *HealthHandler {
	return &HealthHandler{Address: addr}
}

func (h *HealthHandler) Handle() {
	healthRouter := mux.NewRouter()
	healthRouter.HandleFunc("/healthz", livenessHandler())
	http.ListenAndServe(h.Address, healthRouter)
}

func livenessHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	}
}
