package provisioner

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

// Graphqlizer is responsible for converting Go objects to input arguments in graphql format
type graphqlizer struct{}

func (g *graphqlizer) ProvisionRuntimeInputToGraphQL(in gqlschema.ProvisionRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .RuntimeInput }}
		runtimeInput: {{ RuntimeInputToGraphQL .RuntimeInput }}
		{{- end }}
		{{- if .ClusterConfig }}
		clusterConfig: {{ ClusterConfigToGraphQL .ClusterConfig }}
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }}
		{{- end }}
	}`)
}

func (g *graphqlizer) RuntimeInputToGraphQL(in gqlschema.RuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		name: "{{ .Name }}"	
		{{- if .Description }}
		description: {{ .Description }}
		{{- end }}
		{{- if .Labels }}
		labels: {{ .Labels }}
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
		kubernetesVersion: "{{ .KubernetesVersion }}"
		volumeSizeGB: {{ .VolumeSizeGb }}
		machineType: "{{ .MachineType }}"
		region: "{{ .Region }}"
		provider: "{{ .Provider }}"
		diskType: "{{ .DiskType }}"
		{{- if .Seed }}
		seed: "{{ .Seed }}"
		{{- end }}
		targetSecret: "{{ .TargetSecret }}"
		workerCidr: "{{ .WorkerCidr }}"
        autoScalerMin: {{ .AutoScalerMin }}
        autoScalerMax: {{ .AutoScalerMax }}
        maxSurge: {{ .MaxSurge }}
		maxUnavailable: {{ .MaxUnavailable }}
		providerSpecificConfig: {{ ProviderSpecificInputToGraphQL .ProviderSpecificConfig }}
	}`)
}

func (g *graphqlizer) ProviderSpecificInputToGraphQL(in *gqlschema.ProviderSpecificInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .AzureConfig }}
		azureConfig: {{ AzureProviderConfigInputToGraphQL .AzureConfig }}
		{{- end }}
		{{- if .GcpConfig }}
		gcpConfig: {{ GcpProviderConfigInputToGraphQL .GcpConfig }}
		{{- end }}
	}`)
}

func (g *graphqlizer) AzureProviderConfigInputToGraphQL(in *gqlschema.AzureProviderConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		vnetCidr: "{{ .VnetCidr }}"
	}`)
}

func (g *graphqlizer) GcpProviderConfigInputToGraphQL(in *gqlschema.GCPProviderConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		zone: "{{ .Zone }}"
	}`)
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
		version: "{{.Version}}"
		{{- if .Components }}
		components: [
			{{- range $i, $e := .Components }}
			{{- if $i}}, {{- end}} {{ ComponentConfigurationInputToGQL $e }}
			{{- end }}]
		{{- end }}
		{{- if .Configuration }}
		configuration: [
			{{- range $i, $e := .Configuration }}
			{{- if $i}}, {{- end}} {{ ConfigEntryInputToGQL $e }}
			{{- end }}]
		{{- end }}
	}`)
}

func (g *graphqlizer) ComponentConfigurationInputToGQL(in gqlschema.ComponentConfigurationInput) (string, error) {
	return g.genericToGraphQL(in, `{
		component: "{{.Component}}"
		namespace: "{{.Namespace}}"
		{{- if .Configuration }}
		configuration: [
			{{- range $i, $e := .Configuration }}
			{{- if $i}}, {{- end}} {{ ConfigEntryInputToGQL $e }}
			{{- end }}]
		{{- end }}
	}`)
}

func (g *graphqlizer) ConfigEntryInputToGQL(in gqlschema.ConfigEntryInput) (string, error) {
	return g.genericToGraphQL(in, `{
		key: "{{.Key}}"
		value: "{{.Value}}"
		{{- if .Secret }}
		secret: {{.Secret}}
		{{- end }}
	}`)
}

func (g *graphqlizer) genericToGraphQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["RuntimeInputToGraphQL"] = g.RuntimeInputToGraphQL
	fm["ComponentConfigurationInputToGQL"] = g.ComponentConfigurationInputToGQL
	fm["ConfigEntryInputToGQL"] = g.ConfigEntryInputToGQL
	fm["ClusterConfigToGraphQL"] = g.ClusterConfigToGraphQL
	fm["KymaConfigToGraphQL"] = g.KymaConfigToGraphQL
	fm["UpgradeClusterConfigToGraphQL"] = g.UpgradeClusterConfigToGraphQL
	fm["CredentialsInputToGraphQL"] = g.CredentialsInputToGraphQL
	fm["GardenerConfigInputToGraphQL"] = g.GardenerConfigInputToGraphQL
	fm["ProviderSpecificInputToGraphQL"] = g.ProviderSpecificInputToGraphQL
	fm["AzureProviderConfigInputToGraphQL"] = g.AzureProviderConfigInputToGraphQL
	fm["GcpProviderConfigInputToGraphQL"] = g.GcpProviderConfigInputToGraphQL
	fm["GCPConfigInputToGraphQL"] = g.GCPConfigInputToGraphQL

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
