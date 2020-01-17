package runtimes

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/runtimes/clientbuilder"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const configMapName = "compass-agent-configuration"

type RuntimeConfig struct {
	ConnectorURL string
	RuntimeID    string
	Tenant       string
	OneTimeToken string
}

//go:generate mockery -name=ConfigProvider
type ConfigProvider interface {
	CreateConfigMapForRuntime(runtimeConfig RuntimeConfig, kubeconfigRaw string) (*core.ConfigMap, error)
}

type provider struct {
	namespace string
	builder   clientbuilder.ConfigMapClientBuilder
}

func NewRuntimeConfigProvider(namespace string, builder clientbuilder.ConfigMapClientBuilder) ConfigProvider {
	return &provider{
		namespace: namespace,
		builder:   builder,
	}
}

func (p *provider) CreateConfigMapForRuntime(runtimeConfig RuntimeConfig, kubeconfigRaw string) (*core.ConfigMap, error) {
	configMapInterface, err := p.builder.CreateK8SConfigMapClient(kubeconfigRaw, p.namespace)

	if err != nil {
		return nil, err
	}

	configMap := &core.ConfigMap{
		TypeMeta: meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      configMapName,
			Namespace: p.namespace,
		},
		Data: map[string]string{
			"CONNECTOR_URL": runtimeConfig.ConnectorURL,
			"RUNTIME_ID":    runtimeConfig.RuntimeID,
			"TENANT":        runtimeConfig.Tenant,
			"TOKEN":         runtimeConfig.OneTimeToken,
		},
	}
	configMap, err = configMapInterface.Create(configMap)

	return configMap, err
}
