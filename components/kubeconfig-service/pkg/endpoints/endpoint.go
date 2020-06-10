package endpoints

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func GetKubeConfig(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tenant := vars["tenantID"]
	runtime := vars["runtimeID"]

	log.Infof("Fetching kubeconfig for %s/%s", tenant, runtime)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("Error ocurred while reading request data: %s", err)
	}
	log.Infof("%s", body)
}

func GetHealthStatus(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}