package endpoints

import (
	"net/http"

	"github.com/kyma-project/control-plane/components/kubeconfig-service/pkg/transformer"

	"github.com/gorilla/mux"
	"github.com/kyma-project/control-plane/components/kubeconfig-service/pkg/caller"
	log "github.com/sirupsen/logrus"
)

const (
	mimeTypeYaml = "application/x-yaml"
	mimeTypeText = "text/plain"
)

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

	log.Infof("Generating kubeconfig for %s/%s", tenant, runtime)

	kubeConfig, err := ec.generateKubeConfig(tenant, runtime)
	if err != nil {
		w.Header().Add("Content-Type", mimeTypeText)
		w.WriteHeader(http.StatusInternalServerError)
		_, err2 := w.Write([]byte(err.Error()))
		log.Errorf("Error while processing the kubeconfig file: %s", err)
		if err2 != nil {
			log.Errorf("Error while sending response: %s", err2)
		}
	}
	w.Header().Add("Content-Type", mimeTypeYaml)
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

func (ec EndpointClient) generateKubeConfig(tenant, runtime string) ([]byte, error) {
	rawConfig, err := ec.callGQL(tenant, runtime)
	if err != nil || rawConfig == "" {
		return nil, err
	}
	tc, err := transformer.NewClient(rawConfig)
	if err != nil {
		return nil, err
	}
	kubeConfig, err := tc.TransformKubeconfig()
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}
