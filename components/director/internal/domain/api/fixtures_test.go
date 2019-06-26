package api_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

func fixModelAPIDefinition(id, name, description string) *model.APIDefinition {
	return &model.APIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
	}
}

func fixGQLAPIDefinition(id, name, description string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
	}
}

func fixDetailedModelAPIDefinition(t *testing.T, id, name, description string, group string) *model.APIDefinition {
	data := []byte("data")
	format := model.SpecFormatJSON

	spec := &model.APISpec{
		Data:         &data,
		Format:       &format,
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

	auth1 := model.Auth{
		AdditionalHeaders: map[string][]string{"testHeader": {"hval1", "hval2"}},
	}
	auth2 := model.Auth{
		AdditionalQueryParams: map[string][]string{"testParam": {"pval1", "pval2"}},
	}

	runtimeAuth1 := model.RuntimeAuth{
		RuntimeID: "1",
		Auth:      &auth1,
	}
	runtimeAuth2 := model.RuntimeAuth{
		RuntimeID: "2",
		Auth:      &auth2,
	}

	return &model.APIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Auths:         []*model.RuntimeAuth{&runtimeAuth1, &runtimeAuth2},
		DefaultAuth:   &auth1,
		Version:       version,
	}
}

func fixDetailedGQLAPIDefinition(t *testing.T, id, name, description string, group string) *graphql.APIDefinition {
	data := graphql.CLOB("data")
	format := graphql.SpecFormatJSON

	spec := &graphql.APISpec{
		Data:         &data,
		Format:       &format,
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
	params := graphql.QueryParams{"testParam": {"pval1", "pval2"}}

	auth1 := graphql.Auth{
		AdditionalHeaders: &headers,
	}
	auth2 := graphql.Auth{
		AdditionalQueryParams: &params,
	}

	runtimeAuth1 := graphql.RuntimeAuth{
		RuntimeID: "1",
		Auth:      &auth1,
	}
	runtimeAuth2 := graphql.RuntimeAuth{
		RuntimeID: "2",
		Auth:      &auth2,
	}

	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Auth:          nil,
		Auths:         []*graphql.RuntimeAuth{&runtimeAuth1, &runtimeAuth2},
		DefaultAuth:   &auth1,
		Version:       version,
	}
}

func fixModelAPIDefinitionInput(name, description string, group string) *model.APIDefinitionInput {
	data := []byte("data")
	format := model.SpecFormatYaml

	spec := &model.APISpecInput{
		Data:         &data,
		Type:         model.APISpecTypeOpenAPI,
		Format:       &format,
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
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Spec:          spec,
		Version:       version,
		DefaultAuth:   &authInput,
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
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Spec:          spec,
		Version:       version,
		DefaultAuth:   &defaultAuth,
	}
}
