package provisioner

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"reflect"
	"text/template"

	"fmt"

	"strconv"

	"github.com/Masterminds/sprig"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

// Graphqlizer is responsible for converting Go objects to input arguments in graphql format
type Graphqlizer struct{}

func (g *Graphqlizer) ProvisionRuntimeInputToGraphQL(in gqlschema.ProvisionRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .RuntimeInput }}
      	runtimeInput: {{ RuntimeInputToGraphQL .RuntimeInput }},
		{{- end }}
		{{- if .ClusterConfig }}
		clusterConfig: {{ ClusterConfigToGraphQL .ClusterConfig }},
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }},
		{{- end }}
	}`)
}

func (g *Graphqlizer) UpgradeRuntimeInputToGraphQL(in gqlschema.UpgradeRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .ClusterConfig }}
		clusterConfig: {{ UpgradeClusterConfigToGraphQL .ClusterConfig }},
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }},
		{{- end }}
	}`)
}

func (g *Graphqlizer) RuntimeInputToGraphQL(in gqlschema.RuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		{{- if .Labels }}
		labels: {{ LabelsToGQL .Labels}},
		{{- end }}
	}`)
}

func (g *Graphqlizer) LabelsToGQL(in gqlschema.Labels) (string, error) {
	return g.marshal(in), nil
}

func (g *Graphqlizer) ClusterConfigToGraphQL(in gqlschema.ClusterConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .GardenerConfig }}
		gardenerConfig: {{ GardenerConfigInputToGraphQL .GardenerConfig }},
		{{- end }}
		{{- if .GcpConfig }}
		gcpConfig: {{ GCPConfigInputToGraphQL .GcpConfig }},
		{{- end }}
	}`)
}

func (g *Graphqlizer) GardenerConfigInputToGraphQL(in gqlschema.GardenerConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		kubernetesVersion: "{{.KubernetesVersion}}",
		volumeSizeGB: {{.VolumeSizeGb }},
		machineType: "{{.MachineType}}",
		region: "{{.Region}}",
		provider: "{{ .Provider }}",
		diskType: "{{.DiskType}}",
		targetSecret: "{{ .TargetSecret }}",
		workerCidr: "{{ .WorkerCidr }}",
        autoScalerMin: {{ .AutoScalerMin }},
        autoScalerMax: {{ .AutoScalerMax }},
        maxSurge: {{ .MaxSurge }},
		maxUnavailable: {{ .MaxUnavailable }},
		providerSpecificConfig: {
		{{- if .ProviderSpecificConfig.AzureConfig }}
			azureConfig: {{ AzureProviderConfigInputToGraphQL .ProviderSpecificConfig.AzureConfig }},
		{{- end}}
		{{- if .ProviderSpecificConfig.GcpConfig }}
			gcpConfig: {{ GCPProviderConfigInputToGraphQL .ProviderSpecificConfig.GcpConfig }},
		{{- end}}
        }
	}`)
}

func (g *Graphqlizer) AzureProviderConfigInputToGraphQL(in gqlschema.AzureProviderConfigInput) (string, error) {
	return fmt.Sprintf(`{ vnetCidr: "%s" }`, in.VnetCidr), nil
}

func (g *Graphqlizer) GCPProviderConfigInputToGraphQL(in gqlschema.GCPProviderConfigInput) (string, error) {
	return fmt.Sprintf(`{ zone: "%s" }`, in.Zone), nil
}

func (g *Graphqlizer) AWSProviderConfigInputToGraphQL(in gqlschema.AWSProviderConfigInput) (string, error) {
	return fmt.Sprintf(`{ 
		zone: "%s" ,
		publicCidr: "%s",
		vpcCidr: "%s",
        internalCidr: "%s",
}`, in.Zone, in.PublicCidr, in.VpcCidr, in.InternalCidr), nil
}

func (g *Graphqlizer) GCPConfigInputToGraphQL(in gqlschema.GCPConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		name: "{{.Name}}",
		kubernetesVersion: "{{.KubernetesVersion}}",
        projectName: "{{.ProjectName}}",
		numberOfNodes: {{.NumberOfNodes}},
		bootDiskSizeGB: {{ .BootDiskSizeGb }},
		machineType: "{{.MachineType}}",
		region: "{{.Region}}",
		{{- if .Zone }}
		zone: "{{.Zone}}",
		{{- end }}
	}`)
}

func (g *Graphqlizer) UpgradeClusterConfigToGraphQL(in gqlschema.UpgradeClusterInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .Version }}
		version: "{{.Version}}",
		{{- end }}
	}`)
}

func (g *Graphqlizer) KymaConfigToGraphQL(in gqlschema.KymaConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		version: "{{ .Version }}",
      	{{- with .Components }}
        components: [
		  {{- range . }}
          {
            component: "{{ .Component }}",
            namespace: "{{ .Namespace }}",
            {{- if .SourceURL }}
            sourceURL: "{{ .SourceURL }}",
            {{- end }}
      	    {{- with .Configuration }}
            configuration: [
			  {{- range . }}
              {
                key: "{{ .Key }}",
                value: {{ .Value | strQuote }},
				{{- if .Secret }}
                secret: true,
				{{- end }}
              }
		      {{- end }} 
            ]
		    {{- end }} 
          }
		  {{- end }} 
        ]
      	{{- end }}
		{{- with .Configuration }}
		configuration: [
		  {{- range . }}
		  {
			key: "{{ .Key }}",
			value: "{{ .Value }}",
			{{- if .Secret }}
			secret: true,
			{{- end }}
		  }
		  {{- end }}
		]
		{{- end }}
	}`)
}

func (g *Graphqlizer) marshal(obj interface{}) string {
	var out string

	val := reflect.ValueOf(obj)

	switch val.Kind() {
	case reflect.Map:
		s, err := g.genericToGraphQL(obj, `{ {{- range $k, $v := . }}{{ $k }}:{{ marshal $v }},{{ end -}} }`)
		if err != nil {
			logrus.Warnf("failed to marshal labels: %s", err.Error())
			return ""
		}
		out = s
	case reflect.Slice, reflect.Array:
		s, err := g.genericToGraphQL(obj, `[{{ range $i, $e := . }}{{ if $i }},{{ end }}{{ marshal $e }}{{ end }}]`)
		if err != nil {
			return ""
		}
		out = s
	default:
		marshalled, err := json.Marshal(obj)
		if err != nil {
			return ""
		}
		out = string(marshalled)
	}

	return out
}

func (g *Graphqlizer) genericToGraphQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["marshal"] = g.marshal
	fm["RuntimeInputToGraphQL"] = g.RuntimeInputToGraphQL
	fm["ClusterConfigToGraphQL"] = g.ClusterConfigToGraphQL
	fm["KymaConfigToGraphQL"] = g.KymaConfigToGraphQL
	fm["UpgradeClusterConfigToGraphQL"] = g.UpgradeClusterConfigToGraphQL
	fm["GardenerConfigInputToGraphQL"] = g.GardenerConfigInputToGraphQL
	fm["GCPConfigInputToGraphQL"] = g.GCPConfigInputToGraphQL
	fm["AzureProviderConfigInputToGraphQL"] = g.AzureProviderConfigInputToGraphQL
	fm["GCPProviderConfigInputToGraphQL"] = g.GCPProviderConfigInputToGraphQL
	fm["AWSProviderConfigInputToGraphQL"] = g.AWSProviderConfigInputToGraphQL
	fm["LabelsToGQL"] = g.LabelsToGQL
	fm["strQuote"] = strconv.Quote

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
