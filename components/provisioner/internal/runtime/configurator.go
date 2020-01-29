package runtime

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime/clientbuilder"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	configMapName             = "compass-agent-configuration"
	runtimeAgentComponentName = "compass-runtime-agent"
)

//go:generate mockery -name=Configurator
type Configurator interface {
	ConfigureRuntime(cluster model.Cluster, kubeconfigRaw string) error
}

type configurator struct {
	builder        clientbuilder.ConfigMapClientBuilder
	directorClient director.DirectorClient
}

func NewRuntimeConfigurator(builder clientbuilder.ConfigMapClientBuilder, directorClient director.DirectorClient) Configurator {
	return &configurator{
		builder:        builder,
		directorClient: directorClient,
	}
}

func (c *configurator) ConfigureRuntime(cluster model.Cluster, kubeconfigRaw string) error {
	runtimeAgentComponent, found := cluster.KymaConfig.GetComponentConfig(runtimeAgentComponentName)
	if found {
		err := c.configureAgent(cluster, runtimeAgentComponent.Namespace, kubeconfigRaw)
		if err != nil {
			return fmt.Errorf("error configuring Runtime Agent: %s", err.Error())
		}
	}

	return nil
}

func (c *configurator) configureAgent(cluster model.Cluster, namespace, kubeconfigRaw string) error {
	token, err := c.directorClient.GetConnectionToken(cluster.ID, cluster.Tenant)
	if err != nil {
		return fmt.Errorf("error getting one time token from Director: %s", err.Error())
	}

	configMapInterface, err := c.builder.CreateK8SConfigMapClient(kubeconfigRaw, namespace)
	if err != nil {
		return fmt.Errorf("error creating Config Map client: %s", err.Error())
	}

	configMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"CONNECTOR_URL": token.ConnectorURL,
			"RUNTIME_ID":    cluster.ID,
			"TENANT":        cluster.Tenant,
			"TOKEN":         token.Token,
		},
	}

	configMap, err = configMapInterface.Create(configMap)
	if err != nil {
		return fmt.Errorf("error creating Config Map on Runtime: %s", err.Error())
	}

	return nil
}
