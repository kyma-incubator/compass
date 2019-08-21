package eventapi_test

import (

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelEventAPIDefinition(id, appID, name, description string) *model.EventAPIDefinition {
	return &model.EventAPIDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          name,
		Description:   &description,
	}
}

func fixGQLEventAPIDefinition(id, appID, name, description string) *graphql.EventAPIDefinition {
	return &graphql.EventAPIDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          name,
		Description:   &description,
	}
}

func fixDetailedModelEventAPIDefinition(id, name, description string, group string) *model.EventAPIDefinition {
	data := "data"
	format := model.SpecFormatJSON

	frID := "test"
	spec := &model.EventAPISpec{
		Data:         &data,
		Format:       format,
		Type:         model.EventAPISpecTypeAsyncAPI,
		FetchRequestID: &frID,
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

	return &model.EventAPIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Group:         &group,
		Spec:          spec,
		Version:       version,
	}
}

func fixDetailedGQLEventAPIDefinition(id, name, description string, group string) *graphql.EventAPIDefinition {
	data := graphql.CLOB("data")
	format := graphql.SpecFormatJSON

	spec := &graphql.EventAPISpec{
		Data:         &data,
		Format:       format,
		Type:         graphql.EventAPISpecTypeAsyncAPI,
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

	return &graphql.EventAPIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Spec:          spec,
		Group:         &group,
		Version:       version,
	}
}

func fixModelEventAPIDefinitionInput(name, description string, group string) *model.EventAPIDefinitionInput {
	data := "data"
	format := model.SpecFormatYaml

	spec := &model.EventAPISpecInput{
		Data:          &data,
		EventSpecType: model.EventAPISpecTypeAsyncAPI,
		Format:        format,
		FetchRequest:  &model.FetchRequestInput{},
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

	return &model.EventAPIDefinitionInput{
		Name:        name,
		Description: &description,
		Group:       &group,
		Spec:        spec,
		Version:     version,
	}
}

func fixGQLEventAPIDefinitionInput(name, description string, group string) *graphql.EventAPIDefinitionInput {
	data := graphql.CLOB("data")

	spec := &graphql.EventAPISpecInput{
		Data:          &data,
		EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
		Format:        graphql.SpecFormatYaml,
		FetchRequest:  &graphql.FetchRequestInput{},
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

	return &graphql.EventAPIDefinitionInput{
		Name:        name,
		Description: &description,
		Group:       &group,
		Spec:        spec,
		Version:     version,
	}
}
