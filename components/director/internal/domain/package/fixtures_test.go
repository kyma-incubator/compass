package mp_package_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	packageID = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	appID     = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantID  = "ttttttttt-tttt-tttt-tttt-tttttttttttt"
)

func fixPackageModel(t *testing.T, name, desc string) *model.Package {
	return &model.Package{
		ID:                             packageID,
		TenantID:                       tenantID,
		ApplicationID:                  appID,
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(t),
		DefaultInstanceAuth:            fixModelAuth(),
	}
}

func fixGQLPackage(id, name, desc string) *graphql.Package {
	schema := graphql.JSONSchema(`{"$id":"https://example.com/person.schema.json","$schema":"http://json-schema.org/draft-07/schema#","properties":{"age":{"description":"Age in years which must be equal to or greater than zero.","minimum":0,"type":"integer"},"firstName":{"description":"The person's first name.","type":"string"},"lastName":{"description":"The person's last name.","type":"string"}},"title":"Person","type":"object"}`)
	return &graphql.Package{
		ID:                             id,
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: &schema,
		DefaultInstanceAuth:            fixGQLAuth(),
	}
}

func fixGQLPackageCreateInput(name, description string) graphql.PackageCreateInput {
	basicCredentialDataInput := graphql.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}

	credentialDataInput := graphql.CredentialDataInput{Basic: &basicCredentialDataInput}
	defaultAuth := graphql.AuthInput{
		Credential: &credentialDataInput,
	}

	return graphql.PackageCreateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicInputSchema(),
		DefaultInstanceAuth:            &defaultAuth,
	}
}

func fixModelPackageCreateInput(t *testing.T, name, description string) model.PackageCreateInput {
	basicCredentialDataInput := model.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	authInput := model.AuthInput{
		Credential: &model.CredentialDataInput{Basic: &basicCredentialDataInput},
	}

	return model.PackageCreateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicSchema(t),
		DefaultInstanceAuth:            &authInput,
	}
}

func fixGQLPackageUpdateInput(name, description string) graphql.PackageUpdateInput {
	basicCredentialDataInput := graphql.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}

	credentialDataInput := graphql.CredentialDataInput{Basic: &basicCredentialDataInput}
	defaultAuth := graphql.AuthInput{
		Credential: &credentialDataInput,
	}

	return graphql.PackageUpdateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicInputSchema(),
		DefaultInstanceAuth:            &defaultAuth,
	}
}

func fixModelPackageUpdateInput(t *testing.T, name, description string) model.PackageUpdateInput {
	basicCredentialDataInput := model.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	authInput := model.AuthInput{
		Credential: &model.CredentialDataInput{Basic: &basicCredentialDataInput},
	}

	return model.PackageUpdateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicSchema(t),
		DefaultInstanceAuth:            &authInput,
	}
}

func fixPackageCreateInputModel(t *testing.T, name, desc string) model.PackageCreateInput {
	return model.PackageCreateInput{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(t),
		DefaultInstanceAuth:            fixModelAuthInput(map[string][]string{"test": {"foo", "bar"}}),
		APIDefinitions:                 nil,
		EventDefinitions:               nil,
		Documents:                      nil,
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

	auth := model.Auth{
		AdditionalHeaders: map[string][]string{"testHeader": {"hval1", "hval2"}},
	}

	return model.APIDefinition{
		ID:            packageID,
		ApplicationID: appID,
		Tenant:        tenantID,
		Name:          placeholder,
		Description:   str.Ptr("desc_" + placeholder),
		Spec:          spec,
		TargetURL:     fmt.Sprintf("https://%s.com", placeholder),
		Group:         str.Ptr("group_" + placeholder),
		DefaultAuth:   &auth,
		Version:       v,
	}
}

func fixGQLAPIDefinition(id, appId, name, targetURL string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		TargetURL:     targetURL,
	}
}

func fixFullGQLAPIDefinition(placeholder string) *graphql.APIDefinition {
	data := graphql.CLOB("spec_data_" + placeholder)
	format := graphql.SpecFormatYaml

	spec := &graphql.APISpec{
		Data:         &data,
		Format:       format,
		Type:         graphql.APISpecTypeOpenAPI,
		DefinitionID: packageID,
	}

	deprecated := false
	forRemoval := false

	v := &graphql.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	headers := graphql.HttpHeaders{"testHeader": {"hval1", "hval2"}}

	auth := graphql.Auth{
		AdditionalHeaders: &headers,
	}

	return &graphql.APIDefinition{
		ID:            packageID,
		ApplicationID: appID,
		Name:          placeholder,
		Description:   str.Ptr("desc_" + placeholder),
		Spec:          spec,
		TargetURL:     fmt.Sprintf("https://%s.com", placeholder),
		Group:         str.Ptr("group_" + placeholder),
		DefaultAuth:   &auth,
		Version:       v,
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

	basicCredentialDataInput := model.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	authInput := model.AuthInput{
		Credential: &model.CredentialDataInput{Basic: &basicCredentialDataInput},
	}

	return &model.APIDefinitionInput{
		Name:        name,
		Description: &description,
		TargetURL:   "https://test-url.com",
		Group:       &group,
		Spec:        spec,
		Version:     v,
		DefaultAuth: &authInput,
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

	basicCredentialDataInput := graphql.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}

	credentialDataInput := graphql.CredentialDataInput{Basic: &basicCredentialDataInput}
	defaultAuth := graphql.AuthInput{
		Credential: &credentialDataInput,
	}

	return &graphql.APIDefinitionInput{
		Name:        name,
		Description: &description,
		TargetURL:   "https://test-url.com",
		Group:       &group,
		Spec:        spec,
		Version:     v,
		DefaultAuth: &defaultAuth,
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

func fixEntityPackage(id, name, desc string) *mp_package.Entity {
	descSQL := sql.NullString{desc, true}
	schemaSQL := sql.NullString{
		String: inputSchemaString(),
		Valid:  true,
	}
	authSQL := sql.NullString{
		String: `{"Credential":{"Basic":{"Username":"foo","Password":"bar"},"Oauth":null},"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]},"RequestAuth":{"Csrf":{"TokenEndpointURL":"foo.url","Credential":{"Basic":{"Username":"boo","Password":"far"},"Oauth":null},"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]}}}}`,
		Valid:  true,
	}

	return &mp_package.Entity{
		ID:                            id,
		TenantID:                      tenantID,
		ApplicationID:                 appID,
		Name:                          name,
		Description:                   descSQL,
		InstanceAuthRequestJSONSchema: schemaSQL,
		DefaultInstanceAuth:           authSQL,
	}
}

func fixPackageColumns() []string {
	return []string{"id", "tenant_id", "app_id", "name", "description", "instance_auth_request_json_schema", "default_instance_auth"}
}

func fixPackageRow(id, placeholder string) []driver.Value {
	return []driver.Value{id, tenantID, appID, "foo", "bar", fixSchema(), fixDefaultAuth()}
}

func fixPackageCreateArgs(defAuth, schema string, pkg *model.Package) []driver.Value {
	return []driver.Value{packageID, tenantID, appID, pkg.Name, pkg.Description, schema, defAuth}
}

func fixDefaultAuth() string {
	return `{"Credential":{"Basic":{"Username":"foo","Password":"bar"},"Oauth":null},"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]},"RequestAuth":{"Csrf":{"TokenEndpointURL":"foo.url","Credential":{"Basic":{"Username":"boo","Password":"far"},"Oauth":null},"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]}}}}`
}

func inputSchemaString() string {
	return `{"$id":"https://example.com/person.schema.json","$schema":"http://json-schema.org/draft-07/schema#","properties":{"age":{"description":"Age in years which must be equal to or greater than zero.","minimum":0,"type":"integer"},"firstName":{"description":"The person's first name.","type":"string"},"lastName":{"description":"The person's last name.","type":"string"}},"title":"Person","type":"object"}`
}

func fixBasicInputSchema() *graphql.JSONSchema {
	sch := inputSchemaString()
	jsonSchema := graphql.JSONSchema(sch)
	return &jsonSchema
}

func fixBasicSchema(t *testing.T) *interface{} {
	sch := fixBasicInputSchema()
	require.NotNil(t, sch)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(*sch), &obj)
	require.NoError(t, err)
	var objTemp interface{}
	objTemp = obj
	return &objTemp
}

func fixSchema() string {
	return `{"$id":"https://example.com/person.schema.json","$schema":"http://json-schema.org/draft-07/schema#","properties":{"age":{"description":"Age in years which must be equal to or greater than zero.","minimum":0,"type":"integer"},"firstName":{"description":"The person's first name.","type":"string"},"lastName":{"description":"The person's last name.","type":"string"}},"title":"Person","type":"object"}`
}

//func fixSchema(t *testing.T, propertyName, propertyType, propertyDescription, requiredProperty string) *interface{} {
//	sch := fmt.Sprintf(`{
//		"$id": "https://example.com/person.schema.json",
//		"$schema": "http://json-schema.org/draft-07/schema#",
//		"title": "Person",
//		"type": "object",
//		"properties": {
//		  "%s": {
//		    "type": "%s",
//		    "description": "%s"
//		  }
//		},
//		"required": ["%s"]
//	  }`, propertyName, propertyType, propertyDescription, requiredProperty)
//	var obj map[string]interface{}
//
//	err := json.Unmarshal([]byte(sch), &obj)
//	require.NoError(t, err)
//	var objTemp interface{}
//	objTemp = obj
//	return &objTemp
//}
