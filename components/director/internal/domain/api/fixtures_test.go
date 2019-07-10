package api_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelAPIDefinition(id, appId, name, description string) *model.APIDefinition {
	return &model.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
	}
}

func fixGQLAPIDefinition(id, appId, name, description string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
	}
}

func fixDetailedModelAPIDefinition(t *testing.T, id, name, description string, group string) *model.APIDefinition {
	data := "data"
	format := model.SpecFormatJSON

	spec := &model.APISpec{
		Data:         &data,
		Format:       format,
		Type:         model.APISpecTypeOpenAPI,
		FetchRequest: &model.FetchRequest{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	version := &model.Version{
		Value:           "1.0.0",
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}

	auth := model.Auth{
		AdditionalHeaders: map[string][]string{"testHeader": {"hval1", "hval2"}},
	}

	runtimeAuth := model.RuntimeAuth{
		RuntimeID: "1",
		Auth:      &auth,
	}

	return &model.APIDefinition{
		ID:            id,
		ApplicationID: "1",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Auths:         []*model.RuntimeAuth{&runtimeAuth, &runtimeAuth},
		DefaultAuth:   &auth,
		Version:       version,
	}
}

func fixDetailedGQLAPIDefinition(t *testing.T, id, name, description string, group string) *graphql.APIDefinition {
	data := graphql.CLOB("data")
	format := graphql.SpecFormatJSON

	spec := &graphql.APISpec{
		Data:         &data,
		Format:       format,
		Type:         graphql.APISpecTypeOpenAPI,
		FetchRequest: &graphql.FetchRequest{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	version := &graphql.Version{
		Value:           "1.0.0",
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}

	headers := graphql.HttpHeaders{"testHeader": {"hval1", "hval2"}}

	auth := graphql.Auth{
		AdditionalHeaders: &headers,
	}

	runtimeAuth := graphql.RuntimeAuth{
		RuntimeID: "1",
		Auth:      &auth,
	}

	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: "1",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Auth:          nil,
		Auths:         []*graphql.RuntimeAuth{&runtimeAuth, &runtimeAuth},
		DefaultAuth:   &auth,
		Version:       version,
	}
}

func fixModelAPIDefinitionInput(name, description string, group string) *model.APIDefinitionInput {
	data := "data"
	format := model.SpecFormatYaml

	spec := &model.APISpecInput{
		Data:         &data,
		Type:         model.APISpecTypeOpenAPI,
		Format:       format,
		FetchRequest: &model.FetchRequestInput{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	version := &model.VersionInput{
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
		Version:     version,
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

	version := &graphql.VersionInput{
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
		Version:     version,
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

func fixModelRuntimeAuth(id string, auth *model.Auth) *model.RuntimeAuth {
	return &model.RuntimeAuth{
		RuntimeID: id,
		Auth:      auth,
	}
}

func fixGQLRuntimeAuth(id string, auth *graphql.Auth) *graphql.RuntimeAuth {
	return &graphql.RuntimeAuth{
		RuntimeID: id,
		Auth:      auth,
	}
}
