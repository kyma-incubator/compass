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
type graphqlizer struct{}

func (g *graphqlizer) ProvisionRuntimeInputToGraphQL(in gqlschema.ProvisionRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
      runtimeInput: {
        name: "{{ .RuntimeInput.Name }}"
      }
		{{- if .ClusterConfig }}
		clusterConfig: {{ ClusterConfigToGraphQL .ClusterConfig }}
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }}
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

func (g *graphqlizer) GardenerConfigInputToGraphQL(in gqlschema.GardenerConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
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
	return g.genericToGraphQL(in, `{
		version: "{{ .Version }}"
      	{{- with .Components }}
        components: [
		  {{- range . }}
          {
            component: "{{ .Component }}"
            namespace: "{{ .Namespace }}"
      	    {{- with .Configuration }}
            configuration: [
			  {{- range . }}
              {
                key: "{{ .Key }}"
                value: "{{ .Value }}"
				{{- if .Secret }}
                secret: true
				{{- end }}
              }
		      {{- end }} 
            ]
		    {{- end }} 
          }
		  {{- end }} 
        ]
      	{{- end }}         
	}`)
}

func (g *graphqlizer) genericToGraphQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["ClusterConfigToGraphQL"] = g.ClusterConfigToGraphQL
	fm["KymaConfigToGraphQL"] = g.KymaConfigToGraphQL
	fm["UpgradeClusterConfigToGraphQL"] = g.UpgradeClusterConfigToGraphQL
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
