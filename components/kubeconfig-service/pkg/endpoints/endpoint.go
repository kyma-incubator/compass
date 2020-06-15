package endpoints

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/caller"
	log "github.com/sirupsen/logrus"
)

//EndpointClient Wrpper for Endpoints
type EndpointClient struct {
	gqlURL string
}

//NewEndpointClient return new instance of EndpointClient
func NewEndpointClient(gqlURL string) *EndpointClient {
	return &EndpointClient{
		gqlURL: gqlURL,
	}
}

//GetKubeConfig REST Path for Kubeconfig operations
func (ec EndpointClient) GetKubeConfig(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tenant := vars["tenantID"]
	runtime := vars["runtimeID"]

	log.Infof("Fetching kubeconfig for %s/%s", tenant, runtime)
	rawConfig, err := ec.callGQL(tenant, runtime)
	if err != nil || rawConfig == "" {
		log.Errorf("Error ocurred while processing client data: %s", err)
	}
	log.Infof("%s", rawConfig)
}

//GetHealthStatus REST Path for health checks
func (ec EndpointClient) GetHealthStatus(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (ec EndpointClient) callGQL(tenantID, runtimeID string) (string, error) {
	c := caller.NewCaller(ec.gqlURL, tenantID)
	status, err := c.RuntimeStatus(runtimeID)
	if err != nil {
		return "", err
	}
	return *status.RuntimeConfiguration.Kubeconfig, nil
}
