package graphqlizer

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// Graphqlizer is responsible for converting Go objects to input arguments in graphql format
type Graphqlizer struct{}

// ApplicationRegisterInputToGQL missing godoc
func (g *Graphqlizer) ApplicationRegisterInputToGQL(in graphql.ApplicationRegisterInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .ProviderName }}
		providerName: "{{ .ProviderName }}",
		{{- end }}
		{{- if .Description }}
		description: "{{ .Description }}",
		{{- end }}
        {{- if .Labels }}
		labels: {{ LabelsToGQL .Labels}},
		{{- end }}
		{{- if .Webhooks }}
		webhooks: [
			{{- range $i, $e := .Webhooks }} 
				{{- if $i}}, {{- end}} {{ WebhookInputToGQL $e }}
			{{- end }} ],
		{{- end}}
		{{- if .HealthCheckURL }}
		healthCheckURL: "{{ .HealthCheckURL }}",
		{{- end }}
		{{- if .Bundles }} 
		bundles: [
			{{- range $i, $e := .Bundles }} 
				{{- if $i}}, {{- end}} {{- BundleCreateInputToGQL $e }}
			{{- end }} ],
		{{- end }}
		{{- if .IntegrationSystemID }}
		integrationSystemID: "{{ .IntegrationSystemID }}",
		{{- end }}
		{{- if .StatusCondition }}
		statusCondition: {{ .StatusCondition }},
		{{- end }}
		{{- if .BaseURL }}
		baseUrl: "{{ .BaseURL }}"
		{{- end }}
	}`)
}

// ApplicationUpdateInputToGQL missing godoc
func (g *Graphqlizer) ApplicationUpdateInputToGQL(in graphql.ApplicationUpdateInput) (string, error) {
	return g.genericToGQL(in, `{
		{{- if .ProviderName }}
		providerName: "{{ .ProviderName }}",
		{{- end }}
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		{{- if .HealthCheckURL }}
		healthCheckURL: "{{ .HealthCheckURL }}",
		{{- end }}
		{{- if .BaseURL }}
		baseUrl: "{{ .BaseURL }}",
		{{- end }}
		{{- if .IntegrationSystemID }}
		integrationSystemID: "{{ .IntegrationSystemID }}",
		{{- end }}
		{{- if .StatusCondition }}
		statusCondition: {{ .StatusCondition }}
		{{- end }}
	}`)
}

// ApplicationTemplateInputToGQL missing godoc
func (g *Graphqlizer) ApplicationTemplateInputToGQL(in graphql.ApplicationTemplateInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		applicationInput: {{ ApplicationRegisterInputToGQL .ApplicationInput}},
		{{- if .Placeholders }}
		placeholders: [
			{{- range $i, $e := .Placeholders }} 
				{{- if $i}}, {{- end}} {{ PlaceholderDefinitionInputToGQL $e }}
			{{- end }} ],
		{{- end }}
		accessLevel: {{.AccessLevel}},
		{{- if .Webhooks }}
		webhooks: [
			{{- range $i, $e := .Webhooks }} 
				{{- if $i}}, {{- end}} {{ WebhookInputToGQL $e }}
			{{- end }} ],
		{{- end}}	
	}`)
}

// ApplicationTemplateUpdateInputToGQL missing godoc
func (g *Graphqlizer) ApplicationTemplateUpdateInputToGQL(in graphql.ApplicationTemplateUpdateInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		applicationInput: {{ ApplicationRegisterInputToGQL .ApplicationInput}},
		{{- if .Placeholders }}
		placeholders: [
			{{- range $i, $e := .Placeholders }} 
				{{- if $i}}, {{- end}} {{ PlaceholderDefinitionInputToGQL $e }}
			{{- end }} ],
		{{- end }}
		accessLevel: {{.AccessLevel}},
	}`)
}

// DocumentInputToGQL missing godoc
func (g *Graphqlizer) DocumentInputToGQL(in *graphql.DocumentInput) (string, error) {
	return g.genericToGQL(in, `{
		title: "{{.Title}}",
		displayName: "{{.DisplayName}}",
		description: "{{.Description}}",
		format: {{.Format}},
		{{- if .Kind }}
		kind: "{{.Kind}}",
		{{- end}}
		{{- if .Data }}
		data: "{{.Data}}",
		{{- end}}
		{{- if .FetchRequest }}
		fetchRequest: {{- FetchRequesstInputToGQL .FetchRequest }},
		{{- end}}
}`)
}

// FetchRequestInputToGQL missing godoc
func (g *Graphqlizer) FetchRequestInputToGQL(in *graphql.FetchRequestInput) (string, error) {
	return g.genericToGQL(in, `{
		url: "{{.URL}}",
		{{- if .Auth }}
		auth: {{- AuthInputToGQL .Auth }},
		{{- end }}
		{{- if .Mode }}
		mode: {{.Mode}},
		{{- end}}
		{{- if .Filter}}
		filter: "{{.Filter}}",
		{{- end}}
	}`)
}

// CredentialRequestAuthInputToGQL missing godoc
func (g *Graphqlizer) CredentialRequestAuthInputToGQL(in *graphql.CredentialRequestAuthInput) (string, error) {
	return g.genericToGQL(in, `{
		{{- if .Csrf }}
		csrf: {{ CSRFTokenCredentialRequestAuthInputToGQL .Csrf }},
		{{- end }}
	}`)
}

// CredentialDataInputToGQL missing godoc
func (g *Graphqlizer) CredentialDataInputToGQL(in *graphql.CredentialDataInput) (string, error) {
	return g.genericToGQL(in, ` {
			{{- if .Basic }}
			basic: {
				username: "{{ .Basic.Username }}",
				password: "{{ .Basic.Password }}",
			},
			{{- end }}
			{{- if .Oauth }}
			oauth: {
				clientId: "{{ .Oauth.ClientID }}",
				clientSecret: "{{ .Oauth.ClientSecret }}",
				url: "{{ .Oauth.URL }}",
			},
			{{- end }}
	}`)
}

// OneTimeTokenInputToGQL missing godoc
func (g *Graphqlizer) OneTimeTokenInputToGQL(in *graphql.OneTimeTokenInput) (string, error) {
	return g.genericToGQL(in, ` {
			token: "{{ .Token }}",
			{{- if .ConnectorURL }}
			connectorURL: {{ .ConnectorURL }},
			{{- end }}
			used: "{{ .Used }}"
			expiresAt: "{{ .ExpiresAt }}",
			createdAt: "{{ .CreatedAt }}",
			usedAt: "{{ .UsedAt }}",
			{{- if .Raw }}
			raw: "{{ .Raw }}",
			{{- end }}
			{{- if .RawEncoded }}
			rawEncoded: "{{ .RawEncoded }}",
			{{- end }}
			{{- if .Type }}
			type: {{ .Type }},
			{{- end }}
	}`)
}

// CSRFTokenCredentialRequestAuthInputToGQL missing godoc
func (g *Graphqlizer) CSRFTokenCredentialRequestAuthInputToGQL(in *graphql.CSRFTokenCredentialRequestAuthInput) (string, error) {
	in.AdditionalHeadersSerialized = quoteHTTPHeadersSerialized(in.AdditionalHeadersSerialized)
	in.AdditionalQueryParamsSerialized = quoteQueryParamsSerialized(in.AdditionalQueryParamsSerialized)

	return g.genericToGQL(in, `{
			tokenEndpointURL: "{{ .TokenEndpointURL }}",
			{{- if .Credential }}
			credential: {{ CredentialDataInputToGQL .Credential }},
			{{- end }}
			{{- if .AdditionalHeaders }}
			additionalHeaders: {{ HTTPHeadersToGQL .AdditionalHeaders }},
			{{- end }}
			{{- if .AdditionalHeadersSerialized }}
			additionalHeadersSerialized: {{ .AdditionalHeadersSerialized }},
			{{- end }}
			{{- if .AdditionalQueryParams }}
			additionalQueryParams: {{ QueryParamsToGQL .AdditionalQueryParams }},
			{{- end }}
			{{- if .AdditionalQueryParamsSerialized }}
			additionalQueryParamsSerialized: {{ .AdditionalQueryParamsSerialized }},
			{{- end }}
	}`)
}

// AuthInputToGQL missing godoc
func (g *Graphqlizer) AuthInputToGQL(in *graphql.AuthInput) (string, error) {
	in.AdditionalHeadersSerialized = quoteHTTPHeadersSerialized(in.AdditionalHeadersSerialized)
	in.AdditionalQueryParamsSerialized = quoteQueryParamsSerialized(in.AdditionalQueryParamsSerialized)

	return g.genericToGQL(in, `{
		{{- if .Credential }}
		credential: {{ CredentialDataInputToGQL .Credential }},
		{{- end }}
		{{- if .AccessStrategy }}
		accessStrategy: "{{ .AccessStrategy }}",
		{{- end }}
		{{- if .AdditionalHeaders }}
		additionalHeaders: {{ HTTPHeadersToGQL .AdditionalHeaders }},
		{{- end }}
		{{- if .AdditionalHeadersSerialized }}
		additionalHeadersSerialized: {{ .AdditionalHeadersSerialized }},
		{{- end }}
		{{- if .AdditionalQueryParams }}
		additionalQueryParams: {{ QueryParamsToGQL .AdditionalQueryParams}},
		{{- end }}
		{{- if .AdditionalQueryParamsSerialized }}
		additionalQueryParamsSerialized: {{ .AdditionalQueryParamsSerialized }},
		{{- end }}
		{{- if .RequestAuth }}
		requestAuth: {{ CredentialRequestAuthInputToGQL .RequestAuth }},
		{{- end }}
		{{- if .CertCommonName }}
		requestAuth: {{ .CertCommonName }},
		{{- end }}
		{{- if .OneTimeToken }}
		oneTimeToken: {{ OneTimeTokenInputToGQL .OneTimeToken }}
		{{- end }}
	}`)
}

// LabelsToGQL missing godoc
func (g *Graphqlizer) LabelsToGQL(in graphql.Labels) (string, error) {
	return g.marshal(in), nil
}

// HTTPHeadersToGQL missing godoc
func (g *Graphqlizer) HTTPHeadersToGQL(in graphql.HTTPHeaders) (string, error) {
	return g.genericToGQL(in, `{
		{{- range $k,$v := . }}
			{{$k}}: [
				{{- range $i,$j := $v }}
					{{- if $i}},{{- end}}"{{$j}}"
				{{- end }} ],
		{{- end}}
	}`)
}

// QueryParamsToGQL missing godoc
func (g *Graphqlizer) QueryParamsToGQL(in graphql.QueryParams) (string, error) {
	return g.genericToGQL(in, `{
		{{- range $k,$v := . }}
			{{$k}}: [
				{{- range $i,$j := $v }}
					{{- if $i}},{{- end}}"{{$j}}"
				{{- end }} ],
		{{- end}}
	}`)
}

// WebhookInputToGQL missing godoc
func (g *Graphqlizer) WebhookInputToGQL(in *graphql.WebhookInput) (string, error) {
	return g.genericToGQL(in, `{
		type: {{.Type}},
		{{- if .URL }}
		url: "{{.URL }}",
		{{- end }}
		{{- if .Auth }} 
		auth: {{- AuthInputToGQL .Auth }},
		{{- end }}
		{{- if .Mode }} 
		mode: {{.Mode }},
		{{- end }}
		{{- if .CorrelationIDKey }} 
		correlationIdKey: "{{.CorrelationIDKey }}",
		{{- end }}
		{{- if .RetryInterval }} 
		retryInterval: {{.RetryInterval }},
		{{- end }}
		{{- if .Timeout }} 
		timeout: {{.Timeout }},
		{{- end }}
		{{- if .URLTemplate }} 
		urlTemplate: "{{.URLTemplate }}",
		{{- end }}
		{{- if .InputTemplate }} 
		inputTemplate: "{{.InputTemplate }}",
		{{- end }}
		{{- if .HeaderTemplate }} 
		headerTemplate: "{{.HeaderTemplate }}",
		{{- end }}
		{{- if .OutputTemplate }} 
		outputTemplate: "{{.OutputTemplate }}",
		{{- end }}
		{{- if .StatusTemplate }} 
		statusTemplate: "{{.StatusTemplate }}",
		{{- end }}
	}`)
}

// APIDefinitionInputToGQL missing godoc
func (g *Graphqlizer) APIDefinitionInputToGQL(in graphql.APIDefinitionInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{ .Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end}}
		targetURL: "{{.TargetURL}}",
		{{- if .Group }}
		group: "{{.Group}}",
		{{- end }}
		{{- if .Spec }}
		spec: {{- APISpecInputToGQL .Spec }},
		{{- end }}
		{{- if .Version }}
		version: {{- VersionInputToGQL .Version }},
		{{- end}}
	}`)
}

// EventDefinitionInputToGQL missing godoc
func (g *Graphqlizer) EventDefinitionInputToGQL(in graphql.EventDefinitionInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		{{- if .Spec }}
		spec: {{ EventAPISpecInputToGQL .Spec }},
		{{- end }}
		{{- if .Group }}
		group: "{{.Group}}", 
		{{- end }}
		{{- if .Version }}
		version: {{- VersionInputToGQL .Version }},
		{{- end}}
	}`)
}

// EventAPISpecInputToGQL missing godoc
func (g *Graphqlizer) EventAPISpecInputToGQL(in graphql.EventSpecInput) (string, error) {
	in.Data = quoteCLOB(in.Data)
	return g.genericToGQL(in, `{
		{{- if .Data }}
		data: {{.Data}},
		{{- end }}
		type: {{.Type}},
		{{- if .FetchRequest }}
		fetchRequest: {{- FetchRequesstInputToGQL .FetchRequest }},
		{{- end }}
		format: {{.Format}},
	}`)
}

// APISpecInputToGQL missing godoc
func (g *Graphqlizer) APISpecInputToGQL(in graphql.APISpecInput) (string, error) {
	in.Data = quoteCLOB(in.Data)
	return g.genericToGQL(in, `{
		{{- if .Data}}
		data: {{.Data}},
		{{- end}}	
		type: {{.Type}},
		format: {{.Format}},
		{{- if .FetchRequest }}
		fetchRequest: {{- FetchRequesstInputToGQL .FetchRequest }},
		{{- end }}
	}`)
}

// VersionInputToGQL missing godoc
func (g *Graphqlizer) VersionInputToGQL(in graphql.VersionInput) (string, error) {
	return g.genericToGQL(in, `{
		value: "{{.Value}}",
		{{- if .Deprecated }}
		deprecated: {{.Deprecated}},
		{{- end}}
		{{- if .DeprecatedSince }}
		deprecatedSince: "{{.DeprecatedSince}}",
		{{- end}}
		{{- if .ForRemoval }}
		forRemoval: {{.ForRemoval }},
		{{- end }}
	}`)
}

// RuntimeInputToGQL missing godoc
func (g *Graphqlizer) RuntimeInputToGQL(in graphql.RuntimeInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		{{- if .Labels }}
		labels: {{ LabelsToGQL .Labels}},
		{{- end }}
		{{- if .StatusCondition }}
		statusCondition: {{ .StatusCondition }},
		{{- end }}
	}`)
}

// RuntimeContextInputToGQL missing godoc
func (g *Graphqlizer) RuntimeContextInputToGQL(in graphql.RuntimeContextInput) (string, error) {
	return g.genericToGQL(in, `{
		key: "{{.Key}}",
		value: "{{.Value}}",
	}`)
}

// LabelDefinitionInputToGQL missing godoc
func (g *Graphqlizer) LabelDefinitionInputToGQL(in graphql.LabelDefinitionInput) (string, error) {
	return g.genericToGQL(in, `{
		key: "{{.Key}}",
		{{- if .Schema }}
		schema: {{.Schema}},
		{{- end }}
	}`)
}

// LabelFilterToGQL missing godoc
func (g *Graphqlizer) LabelFilterToGQL(in graphql.LabelFilter) (string, error) {
	return g.genericToGQL(in, `{
		key: "{{.Key}}",
		{{- if .Query }}
		query: "{{- js .Query -}}",
		{{- end }}
	}`)
}

// IntegrationSystemInputToGQL missing godoc
func (g *Graphqlizer) IntegrationSystemInputToGQL(in graphql.IntegrationSystemInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
	}`)
}

// PlaceholderDefinitionInputToGQL missing godoc
func (g *Graphqlizer) PlaceholderDefinitionInputToGQL(in graphql.PlaceholderDefinitionInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
	}`)
}

// TemplateValueInputToGQL missing godoc
func (g *Graphqlizer) TemplateValueInputToGQL(in graphql.TemplateValueInput) (string, error) {
	return g.genericToGQL(in, `{
		placeholder: "{{.Placeholder}}"
		value: "{{.Value}}"
	}`)
}

// ApplicationFromTemplateInputToGQL missing godoc
func (g *Graphqlizer) ApplicationFromTemplateInputToGQL(in graphql.ApplicationFromTemplateInput) (string, error) {
	return g.genericToGQL(in, `{
		templateName: "{{.TemplateName}}"
		{{- if .Values }}
		values: [
			{{- range $i, $e := .Values }} 
				{{- if $i}}, {{- end}} {{ TemplateValueInput $e }}
			{{- end }} ],
		{{- end }},
	}`)
}

// BundleCreateInputToGQL missing godoc
func (g *Graphqlizer) BundleCreateInputToGQL(in graphql.BundleCreateInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{ .Name }}"
		{{- if .Description }}
		description: "{{ .Description }}"
		{{- end }}
		{{- if .InstanceAuthRequestInputSchema }}
		instanceAuthRequestInputSchema: {{ .InstanceAuthRequestInputSchema }}
		{{- end }}
		{{- if .DefaultInstanceAuth }}
		defaultInstanceAuth: {{- AuthInputToGQL .DefaultInstanceAuth }}
		{{- end }}
		{{- if .APIDefinitions }}
		apiDefinitions: [
			{{- range $i, $e := .APIDefinitions }}
				{{- if $i}}, {{- end}} {{ APIDefinitionInputToGQL $e }}
			{{- end }}],
		{{- end }}
		{{- if .EventDefinitions }}
		eventDefinitions: [
			{{- range $i, $e := .EventDefinitions }}
				{{- if $i}}, {{- end}} {{ EventDefinitionInputToGQL $e }}
			{{- end }}],
		{{- end }}
		{{- if .Documents }} 
		documents: [
			{{- range $i, $e := .Documents }} 
				{{- if $i}}, {{- end}} {{- DocumentInputToGQL $e }}
			{{- end }} ],
		{{- end }}
	}`)
}

// BundleUpdateInputToGQL missing godoc
func (g *Graphqlizer) BundleUpdateInputToGQL(in graphql.BundleUpdateInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{ .Name }}"
		{{- if .Description }}
		description: "{{ .Description }}"
		{{- end }}
		{{- if .InstanceAuthRequestInputSchema }}
		instanceAuthRequestInputSchema: {{ .InstanceAuthRequestInputSchema }}
		{{- end }}
		{{- if .DefaultInstanceAuth }}
		defaultInstanceAuth: {{- AuthInputToGQL .DefaultInstanceAuth }}
		{{- end }}
	}`)
}

// BundleInstanceAuthStatusInputToGQL missing godoc
func (g *Graphqlizer) BundleInstanceAuthStatusInputToGQL(in graphql.BundleInstanceAuthStatusInput) (string, error) {
	return g.genericToGQL(in, `{
		condition: {{ .Condition }}
		{{- if .Message }}
		message: "{{ .Message }}"
		{{- end }}
		{{- if .Reason }}
		reason: "{{ .Reason }}"
		{{- end }}
	}`)
}

// BundleInstanceAuthRequestInputToGQL missing godoc
func (g *Graphqlizer) BundleInstanceAuthRequestInputToGQL(in graphql.BundleInstanceAuthRequestInput) (string, error) {
	return g.genericToGQL(in, `{
		{{- if .ID }}
		id: "{{ .ID }}"
		{{- end }}		
		{{- if .Context }}
		context: {{ .Context }}
		{{- end }}
		{{- if .InputParams }}
		inputParams: {{ .InputParams }}
		{{- end }}
	}`)
}

// BundleInstanceAuthSetInputToGQL missing godoc
func (g *Graphqlizer) BundleInstanceAuthSetInputToGQL(in graphql.BundleInstanceAuthSetInput) (string, error) {
	return g.genericToGQL(in, `{
		{{- if .Auth }}
		auth: {{- AuthInputToGQL .Auth}}
		{{- end }}
		{{- if .Status }}
		status: {{- BundleInstanceAuthStatusInputToGQL .Status }}
		{{- end }}
	}`)
}

// LabelSelectorInputToGQL missing godoc
func (g *Graphqlizer) LabelSelectorInputToGQL(in graphql.LabelSelectorInput) (string, error) {
	return g.genericToGQL(in, `{
		key: "{{ .Key }}"
		value: "{{ .Value }}"
	}`)
}

// AutomaticScenarioAssignmentSetInputToGQL missing godoc
func (g *Graphqlizer) AutomaticScenarioAssignmentSetInputToGQL(in graphql.AutomaticScenarioAssignmentSetInput) (string, error) {
	return g.genericToGQL(in, `{
		scenarioName: "{{ .ScenarioName }}"
		selector: {{- LabelSelectorInputToGQL .Selector }}
	}`)
}

// WriteTenantsInputToGQL creates tenant input for writeTenants mutation from multiple tenants
func (g *Graphqlizer) WriteTenantsInputToGQL(in []graphql.BusinessTenantMappingInput) (string, error) {
	return g.genericToGQL(in, `
		{{ $n := (len .) }}
		{{ range $i, $tenant := . }}
			{{- if $i }}, {{- end }}
			{
				name: {{ quote $tenant.Name }},
 				externalTenant: {{ quote $tenant.ExternalTenant }},
				{{- if $tenant.Parent }}
				parent: {{ quote $tenant.Parent }},
				{{- end }}
				{{- if $tenant.Region }}
				region: {{ quote $tenant.Region }},
				{{- end }}
				{{- if $tenant.Subdomain }}
				subdomain: {{ quote $tenant.Subdomain }},
				{{- end }}
				type: {{ quote $tenant.Type }},
				provider: {{ quote $tenant.Provider }}
			}
		{{ end }}`)
}

// DeleteTenantsInputToGQL creates tenant input for deleteTenants mutation from multiple external tenant ids
func (g *Graphqlizer) DeleteTenantsInputToGQL(in []graphql.BusinessTenantMappingInput) (string, error) {
	return g.genericToGQL(in, `
		{{ $n := (len .) }}
		{{ range $i, $tenant := . }}
			{{- if $i }}, {{- end }}
			{{ quote $tenant.ExternalTenant }}
		{{ end }}`)
}

// UpdateTenantsInputToGQL creates tenant input for updateTenant
func (g *Graphqlizer) UpdateTenantsInputToGQL(in graphql.BusinessTenantMappingInput) (string, error) {
	return g.genericToGQL(in, `
		{
			name: {{ quote .Name }},
 			externalTenant: {{ quote .ExternalTenant }},
			{{- if .Parent }}
			parent: {{ quote .Parent }},
			{{- end }}
			{{- if .Region }}
			region: {{ quote .Region }},
			{{- end }}
			{{- if .Subdomain }}
			subdomain: {{ quote .Subdomain }},
			{{- end }}
			type: {{ quote .Type }},
			provider: {{ quote .Provider }}
		}`)
}

func (g *Graphqlizer) marshal(obj interface{}) string {
	var out string

	val := reflect.ValueOf(obj)

	switch val.Kind() {
	case reflect.Map:
		s, err := g.genericToGQL(obj, `{ {{- range $k, $v := . }}{{ $k }}:{{ marshal $v }},{{ end -}} }`)
		if err != nil {
			return ""
		}
		out = s
	case reflect.Slice, reflect.Array:
		s, err := g.genericToGQL(obj, `[{{ range $i, $e := . }}{{ if $i }},{{ end }}{{ marshal $e }}{{ end }}]`)
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

func (g *Graphqlizer) genericToGQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["marshal"] = g.marshal
	fm["ApplicationRegisterInputToGQL"] = g.ApplicationRegisterInputToGQL
	fm["DocumentInputToGQL"] = g.DocumentInputToGQL
	fm["FetchRequesstInputToGQL"] = g.FetchRequestInputToGQL
	fm["AuthInputToGQL"] = g.AuthInputToGQL
	fm["LabelsToGQL"] = g.LabelsToGQL
	fm["WebhookInputToGQL"] = g.WebhookInputToGQL
	fm["APIDefinitionInputToGQL"] = g.APIDefinitionInputToGQL
	fm["EventDefinitionInputToGQL"] = g.EventDefinitionInputToGQL
	fm["APISpecInputToGQL"] = g.APISpecInputToGQL
	fm["VersionInputToGQL"] = g.VersionInputToGQL
	fm["HTTPHeadersToGQL"] = g.HTTPHeadersToGQL
	fm["QueryParamsToGQL"] = g.QueryParamsToGQL
	fm["EventAPISpecInputToGQL"] = g.EventAPISpecInputToGQL
	fm["CredentialDataInputToGQL"] = g.CredentialDataInputToGQL
	fm["CSRFTokenCredentialRequestAuthInputToGQL"] = g.CSRFTokenCredentialRequestAuthInputToGQL
	fm["CredentialRequestAuthInputToGQL"] = g.CredentialRequestAuthInputToGQL
	fm["PlaceholderDefinitionInputToGQL"] = g.PlaceholderDefinitionInputToGQL
	fm["TemplateValueInput"] = g.TemplateValueInputToGQL
	fm["BundleInstanceAuthStatusInputToGQL"] = g.BundleInstanceAuthStatusInputToGQL
	fm["BundleCreateInputToGQL"] = g.BundleCreateInputToGQL
	fm["LabelSelectorInputToGQL"] = g.LabelSelectorInputToGQL
	fm["OneTimeTokenInputToGQL"] = g.OneTimeTokenInputToGQL
	fm["quote"] = strconv.Quote

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

func quoteCLOB(in *graphql.CLOB) *graphql.CLOB {
	if in == nil {
		return nil
	}

	quoted := strconv.Quote(string(*in))
	return (*graphql.CLOB)(&quoted)
}

func quoteHTTPHeadersSerialized(in *graphql.HTTPHeadersSerialized) *graphql.HTTPHeadersSerialized {
	if in == nil {
		return nil
	}

	quoted := strconv.Quote(string(*in))
	return (*graphql.HTTPHeadersSerialized)(&quoted)
}

func quoteQueryParamsSerialized(in *graphql.QueryParamsSerialized) *graphql.QueryParamsSerialized {
	if in == nil {
		return nil
	}

	quoted := strconv.Quote(string(*in))
	return (*graphql.QueryParamsSerialized)(&quoted)
}
