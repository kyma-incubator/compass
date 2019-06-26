package api_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/require"
)

func fixModelApiDefinition(id, name, description string) *model.APIDefinition {

	data := []byte("")
	spec := &model.APISpec{
		Data: &data,
		FetchRequest: &model.FetchRequest{
			Mode: model.FetchModeSingle,
			Status: &model.FetchRequestStatus{
				Condition: model.FetchRequestStatusConditionInitial,
			},
		},
	}
	version := &model.Version{}

	return &model.APIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		Version:       version,
	}
}

func fixGQLApiDefinition(id, name, description string) *graphql.APIDefinition {

	data := graphql.CLOB("")
	spec := &graphql.APISpec{
		Data: &data,
		FetchRequest: &graphql.FetchRequest{
			Mode: graphql.FetchModeSingle,
			Status: &graphql.FetchRequestStatus{
				Condition: graphql.FetchRequestStatusConditionInitial,
			},
		},
	}
	version := &graphql.Version{}

	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		Version:       version,
		Auth:          &graphql.RuntimeAuth{},
		Auths:         []*graphql.RuntimeAuth{},
		DefaultAuth:   &graphql.Auth{},
	}
}

func fixModelApiDefinitionWithSpec(id, name string, spec model.APISpec) *model.APIDefinition {
	return &model.APIDefinition{
		ID:          id,
		Name:        name,
		Description: nil,
		Spec:        &spec,
	}
}

func fixDetailedModelApiDefinition(t *testing.T, id, name, description string, group string) *model.APIDefinition {

	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	data := []byte("data")
	filter := "filter"
	format := model.SpecFormatJSON

	spec := &model.APISpec{
		Data:   &data,
		Format: &format,
		Type:   model.APISpecTypeOpenAPI,
		FetchRequest: &model.FetchRequest{
			URL:    "https://test-fetch-request.com",
			Auth:   nil,
			Mode:   model.FetchModeSingle,
			Filter: &filter,
			Status: &model.FetchRequestStatus{
				Condition: model.FetchRequestStatusConditionInitial,
				Timestamp: time,
			},
		},
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
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "test",
				Password: "pwd",
			},
		},
		AdditionalHeaders:     map[string][]string{"testHeader": {"hval1", "hval2"}},
		AdditionalQueryParams: map[string][]string{"testParam": {"pval1", "pval2"}},
		RequestAuth:           &model.CredentialRequestAuth{},
	}
	auth2 := model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "test",
				Password: "pwd",
			},
		},
		AdditionalHeaders:     map[string][]string{"testHeader": {"hval1", "hval2"}},
		AdditionalQueryParams: map[string][]string{"testParam": {"pval1", "pval2"}},
		RequestAuth:           &model.CredentialRequestAuth{},
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
		Auth:          &runtimeAuth1,                                      //TODO: https://github.com/kyma-incubator/compass/issues/67
		Auths:         []*model.RuntimeAuth{&runtimeAuth1, &runtimeAuth2}, //TODO: https://github.com/kyma-incubator/compass/issues/67
		DefaultAuth:   &auth1,                                             //TODO: https://github.com/kyma-incubator/compass/issues/67
		Version:       version,
	}
}

func fixDetailedGQLApiDefinition(t *testing.T, id, name, description string, group string) *graphql.APIDefinition {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	data := graphql.CLOB("data")
	filter := "filter"
	format := graphql.SpecFormatJSON

	spec := &graphql.APISpec{
		Data:   &data,
		Format: &format,
		Type:   graphql.APISpecTypeOpenAPI,
		FetchRequest: &graphql.FetchRequest{
			URL:    "https://test-fetch-request.com",
			Auth:   nil,
			Mode:   graphql.FetchModeSingle,
			Filter: &filter,
			Status: &graphql.FetchRequestStatus{
				Condition: graphql.FetchRequestStatusConditionInitial,
				Timestamp: graphql.Timestamp(time),
			},
		},
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

	headers1 := graphql.HttpHeaders{"testHeader": {"hval1", "hval2"}}
	headers2 := graphql.HttpHeaders{"testHeader": {"hval1", "hval2"}}
	params1 := graphql.QueryParams{"testParam": {"pval1", "pval2"}}
	params2 := graphql.QueryParams{"testParam": {"pval1", "pval2"}}

	auth1 := graphql.Auth{
		Credential: &graphql.BasicCredentialData{
			Username: "test",
			Password: "pwd",
		},
		AdditionalHeaders:     &headers1,
		AdditionalQueryParams: &params1,
		RequestAuth:           &graphql.CredentialRequestAuth{},
	}
	auth2 := graphql.Auth{
		Credential: &graphql.BasicCredentialData{
			Username: "test",
			Password: "pwd",
		},
		AdditionalHeaders:     &headers2,
		AdditionalQueryParams: &params2,
		RequestAuth:           &graphql.CredentialRequestAuth{},
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
		Auth:          &runtimeAuth1,                                        //TODO: https://github.com/kyma-incubator/compass/issues/67
		Auths:         []*graphql.RuntimeAuth{&runtimeAuth1, &runtimeAuth2}, //TODO: https://github.com/kyma-incubator/compass/issues/67
		DefaultAuth:   &auth1,                                               //TODO: https://github.com/kyma-incubator/compass/issues/67
		Version:       version,
	}
}

func fixModelApiDefinitionInput(name, description string, group string) *model.APIDefinitionInput {

	data := []byte("data")
	mode := model.FetchModeSingle
	format := model.SpecFormatYaml
	filter := "filter"

	spec := &model.APISpecInput{
		Data:   &data,
		Type:   model.APISpecTypeOpenAPI,
		Format: &format,
		FetchRequest: &model.FetchRequestInput{
			URL:    "https://test-fetch-request.com",
			Auth:   nil, //TODO: https://github.com/kyma-incubator/compass/issues/67
			Mode:   &mode,
			Filter: &filter,
		},
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

	credentialRequestAuthInput := model.CredentialRequestAuthInput{}
	basicCredentialDataInput := model.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	authInput := model.AuthInput{
		Credential:            &model.CredentialDataInput{Basic: &basicCredentialDataInput},
		AdditionalHeaders:     map[string][]string{"testHeader": {"hval1", "hval2"}},
		AdditionalQueryParams: map[string][]string{"testParam": {"pval1", "pval2"}},
		RequestAuth:           &credentialRequestAuthInput,
	}
	return &model.APIDefinitionInput{
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Spec:          spec,
		Version:       version,
		DefaultAuth:   &authInput, //TODO: https://github.com/kyma-incubator/compass/issues/67
	}
}

func fixGQLApiDefinitionInput(name, description string, group string) *graphql.APIDefinitionInput {

	data := graphql.CLOB("data")
	mode := graphql.FetchModeSingle
	filter := "filter"

	spec := &graphql.APISpecInput{
		Data:   &data,
		Type:   graphql.APISpecTypeOpenAPI,
		Format: graphql.SpecFormatYaml,
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "https://test-fetch-request.com",
			Auth:   nil, //TODO: https://github.com/kyma-incubator/compass/issues/67
			Mode:   &mode,
			Filter: &filter,
		},
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

	headers := graphql.HttpHeaders{"testHeader": {"hval1", "hval2"}}
	params := graphql.QueryParams{"testParam": {"pval1", "pval2"}}
	basicCredentialDataInput := graphql.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	credentialDataInput := graphql.CredentialDataInput{Basic: &basicCredentialDataInput}
	defaultAuth := graphql.AuthInput{
		Credential:            &credentialDataInput,
		AdditionalHeaders:     &headers,
		AdditionalQueryParams: &params,
		RequestAuth:           &graphql.CredentialRequestAuthInput{},
	}

	return &graphql.APIDefinitionInput{
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		TargetURL:     "https://test-url.com",
		Group:         &group,
		Spec:          spec,
		Version:       version,
		DefaultAuth:   &defaultAuth, //TODO: https://github.com/kyma-incubator/compass/issues/67
	}
}
