package testkit

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
		name: "{{.Name}}",
		{{- if .NodeCount }}
		nodeCount: "{{.NodeCount}}"
		{{- end }}
		{{- if .DiskSize }}
		diskSize: "{{.DiskSize}}"
		{{- end }}
		{{- if .MachineType }}
		machineType: "{{.MachineType}}"
		{{- end }}
		{{- if .ComputeZone }}
		computeZone: "{{.ComputeZone}}"
		{{- end }}
		{{- if .Version }}
		version: "{{.Version}}"
		{{- end }}
		{{- if .Credentials }}
		credentials: {{ CredentialsInputToGraphQL .Credentials }}
		{{- end }}
		{{- if .ProviderConfig }}
		providerConfig: {{ ProviderConfigInputToGraphQL .ProviderConfig }}
		{{- end }}
	}`)
}

func (g *graphqlizer) CredentialsInputToGraphQL(in gqlschema.CredentialsInput) (string, error) {
	return g.genericToGraphQL(in, `{
		secretName: "{{.SecretName}}",
	}`)
}

func (g *graphqlizer) ProviderConfigInputToGraphQL(in gqlschema.ProviderConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
        {{- if .GardenerProviderConfig }}
		gardenerProviderConfig: {{ GardenerProviderConfigInputToGraphQL .GardenerProviderConfig }}
        {{- end }}
		{{- if .GcpProviderConfig }}
        gcpProviderConfig: {{ GCPProviderConfigInputToGraphQL .GcpProviderConfig }}
        {{- end }}
        {{- if .AksProviderConfig }}
        aksProviderConfig: {{ AKSProviderConfigInputToGraphQL .AksProviderConfig }}
        {{- end }}
	}`)
}

func (g *graphqlizer) GardenerProviderConfigInputToGraphQL(in gqlschema.GardenerProviderConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
        targetProvider: "{{ .TargetProvider }}"
        targetSecret: "{{ .TargetSecret }}"
		{{- if .AutoScalerMin }}
        autoScalerMin: "{{ .AutoScalerMin }}"
        {{- end }}
		{{- if .AutoScalerMax }}
        autoScalerMax: "{{ .AutoScalerMax }}"
        {{- end }}
		{{- if .MaxSurge }}
        maxSurge: "{{ .MaxSurge }}"
        {{- end }}
		{{- if .MaxUnavailable }}
        maxUnavailable: "{{ .MaxUnavailable }}"
        {{- end }}
        {{- if .AdditionalProperties }}
        additionalProperties: {{ AdditionalPropertiesToGraphQL .AdditionalProperties }}
        {{- end }}
	}`)
}

func (g *graphqlizer) GCPProviderConfigInputToGraphQL(in gqlschema.GCPProviderConfig) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .AdditionalProperties }}
        additionalProperties: {{ AdditionalPropertiesToGraphQL .AdditionalProperties }}
        {{- end }}
	}`)
}

func (g *graphqlizer) AKSProviderConfigInputToGraphQL(in gqlschema.AKSProviderConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .AdditionalProperties }}
        additionalProperties: {{ AdditionalPropertiesToGraphQL .AdditionalProperties }}
        {{- end }}
	}`)
}

func (g *graphqlizer) AdditionalPropertiesToGraphQL(in gqlschema.AdditionalProperties) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- range $k,$v := . }}
			{{$k}}: [
				{{- range $i,$j := $v }}
					{{- if $i}},{{- end}}"{{$j}}"
				{{- end }} ]
		{{- end}}
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
		{{- if .Version }}
		version: "{{.Version}}"
		{{- end }}
		{{- if .Modules }}
		modules: "{{ .Modules }}"
		{{- end }}
	}`)
}

func (g *graphqlizer) genericToGraphQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["ClusterConfigToGraphQL"] = g.ClusterConfigToGraphQL
	fm["KymaConfigToGraphQL"] = g.KymaConfigToGraphQL
	fm["UpgradeClusterConfigToGraphQL"] = g.UpgradeClusterConfigToGraphQL
	fm["CredentialsInputToGraphQL"] = g.CredentialsInputToGraphQL
	fm["ProviderConfigInputToGraphQL"] = g.ProviderConfigInputToGraphQL
	fm["AdditionalPropertiesToGraphQL"] = g.AdditionalPropertiesToGraphQL
	fm["GardenerProviderConfigInputToGraphQL"] = g.GardenerProviderConfigInputToGraphQL
	fm["GCPProviderConfigInputToGraphQL"] = g.GCPProviderConfigInputToGraphQL
	fm["AKSProviderConfigInputToGraphQL"] = g.AKSProviderConfigInputToGraphQL

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
