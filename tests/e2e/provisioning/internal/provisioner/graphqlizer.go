package provisioner

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

// Graphqlizer is responsible for converting Go objects to input arguments in graphql format
type graphqlizer struct{}

func (g *graphqlizer) UpgradeRuntimeInputToGraphQL(in gqlschema.UpgradeRuntimeInput) (string, error) {
	return g.genericToGraphQL(in, `{
		kymaConfig: {{ KymaConfigToGraphQL .KymaConfig }}
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
	fm["ComponentConfigurationInputToGQL"] = g.ComponentConfigurationInputToGQL
	fm["ConfigEntryInputToGQL"] = g.ConfigEntryInputToGQL
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
