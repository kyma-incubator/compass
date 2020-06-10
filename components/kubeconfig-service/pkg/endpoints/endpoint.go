package endpoints

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func GetKubeConfig(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("Error ocurred while reading request data: %s", err)
	}
	log.Infof("%s", body)
}

func GetHealthStatus(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}