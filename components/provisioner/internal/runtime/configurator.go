package runtime

import (
	"github.com/kyma-project/control-plane/components/provisioner/internal/apperrors"
	"github.com/kyma-project/control-plane/components/provisioner/internal/util"

	"github.com/kyma-project/control-plane/components/provisioner/internal/util/k8s"

	"github.com/kyma-project/control-plane/components/provisioner/internal/director"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AgentConfigurationSecretName = "compass-agent-configuration"
	runtimeAgentComponentName    = "compass-runtime-agent"
)

//go:generate mockery -name=Configurator
type Configurator interface {
	ConfigureRuntime(cluster model.Cluster, kubeconfigRaw string) apperrors.AppError
}

type configurator struct {
	builder        k8s.K8sClientProvider
	directorClient director.DirectorClient
}

func NewRuntimeConfigurator(builder k8s.K8sClientProvider, directorClient director.DirectorClient) Configurator {
	return &configurator{
		builder:        builder,
		directorClient: directorClient,
	}
}

func (c *configurator) ConfigureRuntime(cluster model.Cluster, kubeconfigRaw string) apperrors.AppError {
	runtimeAgentComponent, found := cluster.KymaConfig.GetComponentConfig(runtimeAgentComponentName)
	if found {
		err := c.configureAgent(cluster, runtimeAgentComponent.Namespace, kubeconfigRaw)
		if err != nil {
			return err.Append("error configuring Runtime Agent")
		}
	}

	return nil
}

func (c *configurator) configureAgent(cluster model.Cluster, namespace, kubeconfigRaw string) apperrors.AppError {
	var err apperrors.AppError
	token, err := c.directorClient.GetConnectionToken(cluster.ID, cluster.Tenant)
	if err != nil {
		return err.Append("error getting one time token from Director")
	}

	k8sClient, err := c.builder.CreateK8SClient(kubeconfigRaw)
	if err != nil {
		return err.Append("error creating Config Map client")
	}

	configurationData := map[string]string{
		"CONNECTOR_URL": token.ConnectorURL,
		"RUNTIME_ID":    cluster.ID,
		"TENANT":        cluster.Tenant,
		"TOKEN":         token.Token,
	}

	secret := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      AgentConfigurationSecretName,
			Namespace: namespace,
		},
		StringData: configurationData,
	}

	configMap := &core.ConfigMap{
		ObjectMeta: meta.ObjectMeta{
			Name:      AgentConfigurationSecretName,
			Namespace: namespace,
		},
		Data: configurationData,
	}

	// Creating Config Map is deprecated
	// It should be removed when Kyma older than 1.12 is no longer supported
	_, k8serr := k8sClient.CoreV1().ConfigMaps(namespace).Create(configMap)
	if k8serr != nil {
		return util.K8SErrorToAppError(k8serr).Append("error creating Config Map on Runtime")
	}

	_, k8serr = k8sClient.CoreV1().Secrets(namespace).Create(secret)
	if k8serr != nil {
		return util.K8SErrorToAppError(k8serr).Append("error creating Secret on Runtime")
	}

	return nil
}
