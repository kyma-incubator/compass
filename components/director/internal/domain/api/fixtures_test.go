package api_test

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	apiDefID         = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	tenantID         = "ttttttttt-tttt-tttt-tttt-tttttttttttt"
	externalTenantID = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	bundleID         = "ppppppppp-pppp-pppp-pppp-pppppppppppp"
)

func fixAPIDefinitionModel(id string, bundleID string, name, targetURL string) *model.APIDefinition {
	return &model.APIDefinition{
		ID:         id,
		BundleID:   bundleID,
		Title:      name,
		EntryPoint: targetURL,
	}
}

func fixFullAPIDefinitionModel(placeholder string) model.APIDefinition {
	spec := &model.APISpec{
		Data:   str.Ptr("spec_data_" + placeholder),
		Format: model.SpecFormatYaml,
		Type:   model.APISpecTypeOpenAPI,
	}

	deprecated := false
	forRemoval := false

	v := &model.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return model.APIDefinition{
		ID:          apiDefID,
		Tenant:      tenantID,
		BundleID:    bundleID,
		Title:       placeholder,
		Description: str.Ptr("desc_" + placeholder),
		Specs:       spec,
		EntryPoint:  fmt.Sprintf("https://%s.com", placeholder),
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
	}
}

func fixGQLAPIDefinition(id string, bundleId string, name, targetURL string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		ID:         id,
		BundleID:   bundleId,
		Title:      name,
		EntryPoint: targetURL,
	}
}

func fixFullGQLAPIDefinition(placeholder string) *graphql.APIDefinition {
	data := graphql.CLOB("spec_data_" + placeholder)
	format := graphql.SpecFormatYaml

	spec := &graphql.APISpec{
		Data:         &data,
		Format:       format,
		Type:         graphql.APISpecTypeOpenAPI,
		DefinitionID: apiDefID,
	}

	deprecated := false
	forRemoval := false

	v := &graphql.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return &graphql.APIDefinition{
		ID:          apiDefID,
		BundleID:    bundleID,
		Title:       placeholder,
		Description: str.Ptr("desc_" + placeholder),
		Specs:       spec,
		EntryPoint:  fmt.Sprintf("https://%s.com", placeholder),
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
	}
}

func fixModelAPIDefinitionInput(name, description string, group string) *model.APIDefinitionInput {
	data := "data"

	spec := &model.APISpecInput{
		Data:         &data,
		Type:         model.APISpecTypeOpenAPI,
		Format:       model.SpecFormatYaml,
		FetchRequest: &model.FetchRequestInput{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	v := &model.VersionInput{
		Value:           "1.0.0",
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}

	return &model.APIDefinitionInput{
		Title:       name,
		Description: &description,
		EntryPoint:  "https://test-url.com",
		Group:       &group,
		Specs:       spec,
		Version:     v,
	}
}

func fixGQLAPIDefinitionInput(name, description string, group string) *graphql.APIDefinitionInput {
	data := graphql.CLOB("data")

	spec := &graphql.APISpecInput{
		Data:         &data,
		Type:         graphql.APISpecTypeOpenAPI,
		Format:       graphql.SpecFormatYaml,
		FetchRequest: &graphql.FetchRequestInput{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	v := &graphql.VersionInput{
		Value:           "1.0.0",
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}

	return &graphql.APIDefinitionInput{
		Title:       name,
		Description: &description,
		EntryPoint:  "https://test-url.com",
		Group:       &group,
		Spec:        spec,
		Version:     v,
	}
}

func fixModelAuthInput(headers map[string][]string) *model.AuthInput {
	return &model.AuthInput{
		AdditionalHeaders: headers,
	}
}

func fixGQLAuthInput(headers map[string][]string) *graphql.AuthInput {
	httpHeaders := graphql.HttpHeaders(headers)

	return &graphql.AuthInput{
		AdditionalHeaders: &httpHeaders,
	}
}

func fixModelAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
		AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
		AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
		RequestAuth: &model.CredentialRequestAuth{
			Csrf: &model.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: "foo.url",
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "boo",
						Password: "far",
					},
				},
				AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
				AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
			},
		},
	}
}

func fixGQLAuth() *graphql.Auth {
	return &graphql.Auth{
		Credential: &graphql.BasicCredentialData{
			Username: "foo",
			Password: "bar",
		},
		AdditionalHeaders:     &graphql.HttpHeaders{"test": {"foo", "bar"}},
		AdditionalQueryParams: &graphql.QueryParams{"test": {"foo", "bar"}},
		RequestAuth: &graphql.CredentialRequestAuth{
			Csrf: &graphql.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: "foo.url",
				Credential: &graphql.BasicCredentialData{
					Username: "boo",
					Password: "far",
				},
				AdditionalHeaders:     &graphql.HttpHeaders{"test": {"foo", "bar"}},
				AdditionalQueryParams: &graphql.QueryParams{"test": {"foo", "bar"}},
			},
		},
	}
}

func fixModelAPIRtmAuth(id string, auth *model.Auth) *model.APIRuntimeAuth {
	return &model.APIRuntimeAuth{
		ID:        str.Ptr("foo"),
		TenantID:  "tnt",
		RuntimeID: id,
		APIDefID:  "api_id",
		Value:     auth,
	}
}

func fixEntityAPIDefinition(id string, bundleID string, name, targetUrl string) api.Entity {
	return api.Entity{
		ID:         id,
		BundleID:   bundleID,
		Title:      name,
		EntryPoint: targetUrl,
	}
}

func fixFullEntityAPIDefinition(apiDefID, placeholder string) api.Entity {
	boolPlaceholder := false

	return api.Entity{
		ID:          apiDefID,
		TenantID:    tenantID,
		BundleID:    bundleID,
		Title:       placeholder,
		Description: repo.NewValidNullableString("desc_" + placeholder),
		Group:       repo.NewValidNullableString("group_" + placeholder),
		EntryPoint:  fmt.Sprintf("https://%s.com", placeholder),
		EntitySpec: api.EntitySpec{
			SpecData:   repo.NewValidNullableString("spec_data_" + placeholder),
			SpecFormat: repo.NewValidNullableString(string(model.SpecFormatYaml)),
			SpecType:   repo.NewValidNullableString(string(model.APISpecTypeOpenAPI)),
		},
		Version: version.Version{
			VersionValue:           repo.NewNullableString(str.Ptr("v1.1")),
			VersionDepracated:      repo.NewNullableBool(&boolPlaceholder),
			VersionDepracatedSince: repo.NewNullableString(str.Ptr("v1.0")),
			VersionForRemoval:      repo.NewNullableBool(&boolPlaceholder),
		},
	}
}

func fixAPIDefinitionColumns() []string {
	return []string{"id", "tenant_id", "bundle_id", "title", "description", "group_name", "entry_point",
		//"api_definitions", "tags", "documentation", "changelog_entries", "logo", "image", "url", "release_status", "api_protocol", "actions", "last_updated", "extensions",
		"spec_data", "spec_format", "spec_type", "version", "version_deprecated", "version_deprecated_since", "version_for_removal"}
}

func fixAPIDefinitionRow(id, placeholder string) []driver.Value {
	return []driver.Value{id, tenantID, bundleID, placeholder, "desc_" + placeholder, "group_" + placeholder,
		fmt.Sprintf("https://%s.com", placeholder), "spec_data_" + placeholder, "YAML", "OPEN_API",
		"v1.1", false, "v1.0", false}
}

func fixAPICreateArgs(id string, api *model.APIDefinition) []driver.Value {
	return []driver.Value{id, tenantID, bundleID, api.Title, api.Description, api.Group,
		api.EntryPoint, api.Specs.Data, string(api.Specs.Format), string(api.Specs.Type),
		api.Version.Value, api.Version.Deprecated, api.Version.DeprecatedSince,
		api.Version.ForRemoval}
}

func fixDefaultAuth() string {
	return `{"Credential":{"Basic":null,"Oauth":null},"AdditionalHeaders":{"testHeader":["hval1","hval2"]},"AdditionalQueryParams":null,"RequestAuth":null}`
}

func fixModelFetchRequest(id, url string, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     id,
		Tenant: tenantID,
		URL:    url,
		Auth:   nil,
		Mode:   "SINGLE",
		Filter: nil,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.APIFetchRequestReference,
		ObjectID:   "foo",
	}
}

func fixModelFetchRequestWithCondition(id, url string, timestamp time.Time, condition model.FetchRequestStatusCondition) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     id,
		Tenant: tenantID,
		URL:    url,
		Auth:   nil,
		Mode:   "SINGLE",
		Filter: nil,
		Status: &model.FetchRequestStatus{
			Condition: condition,
			Timestamp: timestamp,
		},
		ObjectType: model.APIFetchRequestReference,
		ObjectID:   "foo",
	}
}

func fixGQLFetchRequest(url string, timestamp time.Time) *graphql.FetchRequest {
	return &graphql.FetchRequest{
		Filter: nil,
		Mode:   graphql.FetchModeSingle,
		Auth:   nil,
		URL:    url,
		Status: &graphql.FetchRequestStatus{
			Timestamp: graphql.Timestamp(timestamp),
			Condition: graphql.FetchRequestStatusConditionInitial,
		},
	}
}
