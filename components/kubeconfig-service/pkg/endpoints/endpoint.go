package endpoints

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/transformer"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/caller"
	log "github.com/sirupsen/logrus"
)

const mimeTypeYaml = "application/x-yaml"

//EndpointClient Wrpper for Endpoints
type EndpointClient struct {
	gqlURL           string
	oidcIssuerURL    string
	oidcClientID     string
	oidcClientSecret string
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

	w.Header().Add("Content-Type", mimeTypeYaml)

	//TODO: Business logic is mixed with low-level HTTP things. This makes testing/maintenance harder and can be easily fixed.
	log.Infof("Fetching kubeconfig for %s/%s", tenant, runtime)
	//TODO: What if tenant/runtime is invalid or not-found? Perhaps a 400/404 error should be returned?
	rawConfig, err := ec.callGQL(tenant, runtime)
	if err != nil || rawConfig == "" {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Error ocurred while processing client data: %s", err)
	}

	tc, err := transformer.NewClient(rawConfig)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Error while decoding kubeconfig from server: %s", err)
	}

	kubeConfig, err := tc.TransformKubeconfig()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Error while processing the kubeconfig file: %s", err)
	}
	//BUG: This is executed even if an error occurred
	log.Infof("Generated new Kubeconfig for %s/%s", tenant, runtime)

	//TODO: In case of an error, we could serialize it's description to YAML/JSON and send that instead.
	_, err = w.Write(kubeConfig)
	if err != nil {
		log.Errorf("Error while sending response: %s", err)
	}
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
