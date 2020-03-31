package health

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Address string
	Log     *log.Logger
}

func NewServer(host, port string, log *log.Logger) *Server {
	return &Server{
		Address: fmt.Sprintf("%s:%s", host, port),
		Log:     log,
	}
}

func (srv *Server) ServeAsync() {
	healthRouter := mux.NewRouter()
	healthRouter.HandleFunc("/healthz", livenessHandler())
	go func() {
		err := http.ListenAndServe(srv.Address, healthRouter)
		if err != nil {
			srv.Log.Errorf("HTTP Health server ListenAndServe: %v", err)
		}
	}()
}

func livenessHandler() func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	}
}
