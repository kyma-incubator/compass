package provisioner

import (
	"bytes"
	"text/template"

	"fmt"

	"github.com/Masterminds/sprig"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

// Graphqlizer is responsible for converting Go objects to input arguments in graphql format
type graphqlizer struct {
	smURL      string
	smUsername string
	smPassword string
}

func (g *graphqlizer) ProvisionRuntimeInputToGraphQL(in gqlschema.ProvisionRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
      runtimeInput: {
        name: "{{ .ClusterConfig.GardenerConfig.Name }}"
      }
		{{- if .ClusterConfig }}
		clusterConfig: {{ ClusterConfigToGraphQL .ClusterConfig }}
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }}
		{{- end }}
		{{- if .Credentials }}
		credentials: {{ CredentialsInputToGraphQL .Credentials }}
		{{- end }}
	}`)
}

func (g *graphqlizer) UpgradeRuntimeInputToGraphQL(in gqlschema.UpgradeRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .ClusterConfig }}
		clusterConfig: {{ UpgradeClusterConfigToGraphQL .ClusterConfig }}
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }}
		{{- end }}
	}`)
}

func (g *graphqlizer) ClusterConfigToGraphQL(in gqlschema.ClusterConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .GardenerConfig }}
		gardenerConfig: {{ GardenerConfigInputToGraphQL .GardenerConfig }}
		{{- end }}
		{{- if .GcpConfig }}
		gcpConfig: {{ GCPConfigInputToGraphQL .GcpConfig }}
		{{- end }}
	}`)
}

func (g *graphqlizer) CredentialsInputToGraphQL(in gqlschema.CredentialsInput) (string, error) {
	return g.genericToGraphQL(in, `{
		secretName: "{{.SecretName}}",
	}`)
}

func (g *graphqlizer) GardenerConfigInputToGraphQL(in gqlschema.GardenerConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		name: "{{.Name}}"
        projectName: "{{.ProjectName}}"
		kubernetesVersion: "{{.KubernetesVersion}}"
		nodeCount: {{.NodeCount}}
		volumeSizeGB: {{.VolumeSizeGb }}
		machineType: "{{.MachineType}}"
		region: "{{.Region}}"
		provider: "{{ .Provider }}"
		diskType: "{{.DiskType}}"
		seed: "az-eu3"
		targetSecret: "{{ .TargetSecret }}"
		workerCidr: "{{ .WorkerCidr }}"
        autoScalerMin: {{ .AutoScalerMin }}
        autoScalerMax: {{ .AutoScalerMax }}
        maxSurge: {{ .MaxSurge }}
		maxUnavailable: {{ .MaxUnavailable }}
		providerSpecificConfig: {
		{{- if .ProviderSpecificConfig.AzureConfig }}
			azureConfig: {{ AzureProviderConfigInputToGraphQL .ProviderSpecificConfig.AzureConfig }}
		{{- end}}
		{{- if .ProviderSpecificConfig.GcpConfig }}
			gcpConfig: {{ GCPProviderConfigInputToGraphQL .ProviderSpecificConfig.GcpConfig }}
		{{- end}}
		{{- if .ProviderSpecificConfig.AwsConfig }}
			awsConfig: {{ AWSProviderConfigInputToGraphQL .ProviderSpecificConfig.AwsConfig }}
		{{- end}}

        }
	}`)
}

func (g *graphqlizer) AzureProviderConfigInputToGraphQL(in gqlschema.AzureProviderConfigInput) (string, error) {
	return fmt.Sprintf(`{ vnetCidr: "%s" }`, in.VnetCidr), nil
}

func (g *graphqlizer) GCPProviderConfigInputToGraphQL(in gqlschema.GCPProviderConfigInput) (string, error) {
	return fmt.Sprintf(`{ zone: "%s" }`, in.Zone), nil
}

func (g *graphqlizer) AWSProviderConfigInputToGraphQL(in gqlschema.AWSProviderConfigInput) (string, error) {
	return fmt.Sprintf(`{ 
		zone: "%s" 
		publicCidr: "%s"
		vpcCidr: "%s"
        internalCidr: "%s"
}`, in.Zone, in.PublicCidr, in.VpcCidr, in.InternalCidr), nil
}

func (g *graphqlizer) GCPConfigInputToGraphQL(in gqlschema.GCPConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		name: "{{.Name}}"
		kubernetesVersion: "{{.KubernetesVersion}}"
        projectName: "{{.ProjectName}}"
		numberOfNodes: {{.NumberOfNodes}}
		bootDiskSizeGB: {{ .BootDiskSizeGb }}
		machineType: "{{.MachineType}}"
		region: "{{.Region}}"
		{{- if .Zone }}
		zone: "{{.Zone}}"
		{{- end }}
	}`)
}

func (g *graphqlizer) UpgradeClusterConfigToGraphQL(in gqlschema.UpgradeClusterInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .Version }}
		version: "{{.Version}}"
		{{- end }}
	}`)
}

func (g *graphqlizer) KymaConfigToGraphQL(in gqlschema.KymaConfigInput) (string, error) {
	return g.genericToGraphQL(in, fmt.Sprintf(`{
		{{- if .Version }}
		version: "master-ece6e5d9"
        components: [
          {
            component: "cluster-essentials"
            namespace: "kyma-system"
          }
          {
            component: "testing"
            namespace: "kyma-system"
          }
          {
            component: "istio-init"
            namespace: "istio-system"
          }
          {
            component: "istio"
            namespace: "istio-system"
          }
          {
            component: "xip-patch"
            namespace: "kyma-installer"
          }
          {
            component: "istio-kyma-patch"
            namespace: "istio-system"
          }
          {
            component: "knative-serving-init"
            namespace: "knative-serving"
          }
          {
            component: "knative-serving"
            namespace: "knative-serving"
          }
          {
            component: "knative-eventing"
            namespace: "knative-eventing"
          }
          {
            component: "dex"
            namespace: "kyma-system"
          }
          {
            component: "ory"
            namespace: "kyma-system"
          }
          {
            component: "api-gateway"
            namespace: "kyma-system"
          }
          {
            component: "service-catalog"
            namespace: "kyma-system"
          }
          {
            component: "service-catalog-addons"
            namespace: "kyma-system"
          }
          {
            component: "rafter"
            namespace: "kyma-system"
          }
          {
            component: "helm-broker"
            namespace: "kyma-system"
          }
          {
            component: "nats-streaming"
            namespace: "natss"
          }
          {
            component: "core"
            namespace: "kyma-system"
          }
          {
            component: "knative-provisioner-natss"
            namespace: "knative-eventing"
          }
          {
            component: "event-bus"
            namespace: "kyma-system"
          }
          {
            component: "event-sources"
            namespace: "kyma-system"
          }
          {
            component: "application-connector-ingress"
            namespace: "kyma-system"
          }    
          {
            component: "application-connector-helper"
            namespace: "kyma-integration"
          }    
          {
            component: "application-connector"
            namespace: "kyma-integration"
          }    
          {
            component: "backup-init"
		    namespace: "kyma-system"
		  }
          {
            component: "backup"
		    namespace: "kyma-system"
		  }
          {
            component: "logging"
		    namespace: "kyma-system"
		  }
          {
            component: "jaeger"
		    namespace: "kyma-system"
		  }
          {
            component: "monitoring"
		    namespace: "kyma-system"
		  }
          {
            component: "kiali"
		    namespace: "kyma-system"
		  }
          {
            component: "service-manager-proxy"
            namespace: "kyma-system"
            configuration: [
              {
                key: "config.sm.url"
                value: "%s"
              }
              {
                key: "sm.password"
                value: "%s"
                secret: true
              }
              {
                key: "sm.user"
                value: "%s"
              }
            ]
          }    
          {
            component: "uaa-activator"
            namespace: "kyma-system"
          }    
          {
            component: "compass-runtime-agent"
            namespace: "compass-system"
          }         
        ]
		{{- end }}
	}`, g.smURL, g.smPassword, g.smUsername))
}

func (g *graphqlizer) genericToGraphQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["ClusterConfigToGraphQL"] = g.ClusterConfigToGraphQL
	fm["KymaConfigToGraphQL"] = g.KymaConfigToGraphQL
	fm["UpgradeClusterConfigToGraphQL"] = g.UpgradeClusterConfigToGraphQL
	fm["CredentialsInputToGraphQL"] = g.CredentialsInputToGraphQL
	fm["GardenerConfigInputToGraphQL"] = g.GardenerConfigInputToGraphQL
	fm["GCPConfigInputToGraphQL"] = g.GCPConfigInputToGraphQL
	fm["AzureProviderConfigInputToGraphQL"] = g.AzureProviderConfigInputToGraphQL
	fm["GCPProviderConfigInputToGraphQL"] = g.GCPProviderConfigInputToGraphQL
	fm["AWSProviderConfigInputToGraphQL"] = g.AWSProviderConfigInputToGraphQL

	t, err := template.New("tmpl").Funcs(fm).Parse(tmpl)
	if err != nil {
		return "", errors.Wrapf(err, "while parsing template")
	}

	var b bytes.Buffer

	if err := t.Execute(&b, obj); err != nil {
		return "", errors.Wrap(err, "while executing template")
	}
	return b.String(), nil
}
