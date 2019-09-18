package testkit

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"text/template"
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
		clusterConfig: {{ ClusterConfigToGraphQL .ClusterConfig }}
		{{- end }}
		{{- if .KymaConfig }}
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }}
		{{- end }}
	}`)
}

func (g *graphqlizer) RuntimeIDInputToGraphQL(in gqlschema.RuntimeIDInput) (string, error) {
	return g.genericToGraphQL(in, `{
		id: .ID
	}`)
}

func (g *graphqlizer) AsyncOperationIDInputToGraphQL(in gqlschema.AsyncOperationIDInput) (string, error) {
	return g.genericToGraphQL(in, `{
		id: .ID
	}`)
}

func (g *graphqlizer) ClusterConfigToGraphQL(in gqlschema.ClusterConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		name: "{{.Name}}",
		{{- if .Size }}
		size: {{.Size}}
		{{- end }}
		{{- if .Memory }}
		memory: {{.Memory}}
		{{- end }}
		{{- if .ComputeZone }}
		computeZone: {{.ComputeZone}}
		{{- end }}
		{{- if .Version }}
		version: {{.Version}}
		{{- end }}
		{{- if .Credentials }}
		credentials: {{.Credentials}}
		{{- end }}
		{{- if .InfrastructureProvider }}
		infrastructureProvider: {{.InfrastructureProvider}}
		{{- end }}
	}`)
}

func (g *graphqlizer) KymaConfigToGraphQL(in gqlschema.KymaConfigInput) (string, error) {
	return g.genericToGraphQL(in, `{
		{{- if .Version }}
		version: {{.Version}}
		{{- end }}
		{{- if .Modules }}
		modules: {{ .Modules }}
		{{- end }}
	}`)
}

func (g *graphqlizer) genericToGraphQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["ClusterConfigToGraphQL"] = g.ClusterConfigToGraphQL
	fm["KymaConfigToGraphQL"] = g.KymaConfigToGraphQL

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
