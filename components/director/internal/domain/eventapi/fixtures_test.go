package eventapi_test


import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelEventAPIDefinition(id, name, description string) *model.EventAPIDefinition {
	return &model.EventAPIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
	}
}

func fixGQLEventAPIDefinition(id, name, description string) *graphql.EventAPIDefinition {
	return &graphql.EventAPIDefinition{
		ID:            id,
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
	}
}

func fixDetailedModelEventAPIDefinition(t *testing.T, id, name, description string, group string) *model.EventAPIDefinition {
	data := []byte("data")
	format := model.SpecFormatJSON

	spec := &model.EventAPISpec{
		Data:         &data,
		Format:       &format,
		Type:         model.EventAPISpecTypeAsyncAPI,
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

func fixDetailedGQLEventAPIDefinition(t *testing.T, id, name, description string, group string) *graphql.EventAPIDefinition {
	data := graphql.CLOB("data")
	format := graphql.SpecFormatJSON

	spec := &graphql.EventAPISpec{
		Data:         &data,
		Format:       &format,
		Type:         graphql.EventAPISpecTypeAsyncAPI,
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

func fixModelAPIEventDefinitionInput(name, description string, group string) *model.EventAPIDefinitionInput {
	data := []byte("data")
	//format := model.SpecFormatYaml

	spec := &model.EventAPISpecInput{
		Data:         &data,
		EventSpecType:         model.EventAPISpecTypeAsyncAPI,
		//Format:       &format, //TODO format
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

	return &model.EventAPIDefinitionInput{
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Group:         &group,
		Spec:          spec,
		Version:       version,
	}
}

func fixGQLAPIEventDefinitionInput(name, description string, group string) *graphql.EventAPIDefinitionInput {
	data := graphql.CLOB("data")

	spec := &graphql.EventAPISpecInput{
		Data:         &data,
		EventSpecType:         graphql.EventAPISpecTypeAsyncAPI,
		//Format:       graphql.SpecFormatYaml, //TODO format
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

	return &graphql.EventAPIDefinitionInput{
		ApplicationID: "applicationID",
		Name:          name,
		Description:   &description,
		Group:         &group,
		Spec:          spec,
		Version:       version,
	}
}

