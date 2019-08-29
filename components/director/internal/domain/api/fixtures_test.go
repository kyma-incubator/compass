package api_test

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

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

func fixDetailedModelAPIDefinition(id, name, description string, group string) *model.APIDefinition {
	data := "data"
	format := model.SpecFormatJSON

	spec := &model.APISpec{
		Data:           &data,
		Format:         format,
		Type:           model.APISpecTypeOpenAPI,
		FetchRequestID: nil,
	}

	deprecated := false
	deprecatedSince := "1.0"
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

func fixDetailedGQLAPIDefinition(id, name, description string, group string) *graphql.APIDefinition {
	data := graphql.CLOB("data")
	format := graphql.SpecFormatJSON

	spec := &graphql.APISpec{
		Data:         &data,
		Format:       format,
		Type:         graphql.APISpecTypeOpenAPI,
		DefinitionID: id,
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

func fixDetailedApiDefinitionEntity(placeholder string) *api.APIDefinition {
	defaultAuthJson := `{"Credential":{"Basic":null,"Oauth":null},"AdditionalHeaders":{"testHeader":["hval1","hval2"]},
							"AdditionalQueryParams":null,"RequestAuth":null}`
	fetchRequestID := uuid.New().String()
	boolPlaceholder := true

	entity := api.APIDefinition{
		ID:          uuid.New().String(),
		TenantID:    uuid.New().String(),
		AppID:       uuid.New().String(),
		Name:        placeholder,
		Description: repo.NewNullableString(&placeholder),
		Group:       repo.NewNullableString(&placeholder),
		TargetURL:   placeholder,
		APISpec: &api.APISpec{
			SpecData:   repo.NewNullableString(&placeholder),
			SpecFormat: repo.NewNullableString(strings.Ptr(string(model.SpecFormatYaml))),
			SpecType:   repo.NewNullableString(strings.Ptr(string(model.APISpecTypeOpenAPI))),
		},
		DefaultAuth:        repo.NewNullableString(&defaultAuthJson),
		SpecFetchRequestID: repo.NewNullableString(&fetchRequestID),
		Version: &version.Version{
			VersionValue:           repo.NewNullableString(&placeholder),
			VersionDepracated:      repo.NewNullableBool(&boolPlaceholder),
			VersionDepracatedSince: repo.NewNullableString(&placeholder),
			VersionForRemoval:      repo.NewNullableBool(&boolPlaceholder),
		},
	}

	return &entity
}

func fixMinimalApiDefinitionEntity(id, app_id, name, targetUrl string) *api.APIDefinition {
	return &api.APIDefinition{
		ID:        id,
		AppID:     app_id,
		Name:      name,
		TargetURL: targetUrl,
	}
}

func fixModelFetchRequest(id, url string, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     id,
		Tenant: "tenant",
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
