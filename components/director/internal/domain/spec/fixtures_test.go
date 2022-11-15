package spec_test

import (
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	specID         = "specID"
	tenant         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	apiID          = "apiID"
	eventID        = "eventID"
	externalTenant = "externalTenant"
)

func fixModelAPISpec() *model.Spec {
	return fixModelAPISpecWithID(specID)
}

func fixModelAPISpecWithID(id string) *model.Spec {
	var specData = "specData"
	var apiType = model.APISpecTypeOdata
	return &model.Spec{
		ID:         id,
		ObjectType: model.APISpecReference,
		ObjectID:   apiID,
		APIType:    &apiType,
		Format:     model.SpecFormatXML,
		Data:       &specData,
	}
}

func fixModelAPISpecWithIDs(id, apiDefID string) *model.Spec {
	var specData = "specData"
	var apiType = model.APISpecTypeOdata
	return &model.Spec{
		ID:         id,
		ObjectType: model.APISpecReference,
		ObjectID:   apiDefID,
		APIType:    &apiType,
		Format:     model.SpecFormatXML,
		Data:       &specData,
	}
}

func fixModelEventSpec() *model.Spec {
	var specData = "specData"
	var eventType = model.EventSpecTypeAsyncAPI
	return &model.Spec{
		ID:         specID,
		ObjectType: model.EventSpecReference,
		ObjectID:   eventID,
		EventType:  &eventType,
		Format:     model.SpecFormatJSON,
		Data:       &specData,
	}
}

func fixModelEventSpecWithID(id string) *model.Spec {
	var specData = "specData"
	var eventType = model.EventSpecTypeAsyncAPI
	return &model.Spec{
		ID:         id,
		ObjectType: model.EventSpecReference,
		ObjectID:   eventID,
		EventType:  &eventType,
		Format:     model.SpecFormatJSON,
		Data:       &specData,
	}
}

func fixModelEventSpecWithIDs(id, eventDefID string) *model.Spec {
	var specData = "specData"
	var eventType = model.EventSpecTypeAsyncAPI
	return &model.Spec{
		ID:         id,
		ObjectType: model.EventSpecReference,
		ObjectID:   eventDefID,
		EventType:  &eventType,
		Format:     model.SpecFormatJSON,
		Data:       &specData,
	}
}

func fixGQLAPISpec() *graphql.APISpec {
	var specData = "specData"
	clob := graphql.CLOB(specData)
	return &graphql.APISpec{
		ID:           specID,
		Data:         &clob,
		DefinitionID: apiID,
		Format:       graphql.SpecFormatXML,
		Type:         graphql.APISpecTypeOdata,
	}
}

func fixGQLEventSpec() *graphql.EventSpec {
	var specData = "specData"
	clob := graphql.CLOB(specData)
	return &graphql.EventSpec{
		ID:           specID,
		Data:         &clob,
		DefinitionID: eventID,
		Format:       graphql.SpecFormatJSON,
		Type:         graphql.EventSpecTypeAsyncAPI,
	}
}

func fixModelAPISpecInput() *model.SpecInput {
	var specData = "specData"
	var apiType = model.APISpecTypeOdata
	return &model.SpecInput{
		Data:    &specData,
		Format:  model.SpecFormatXML,
		APIType: &apiType,
	}
}

func fixModelEventSpecInput() *model.SpecInput {
	var specData = "specData"
	var eventType = model.EventSpecTypeAsyncAPI
	return &model.SpecInput{
		Data:      &specData,
		Format:    model.SpecFormatJSON,
		EventType: &eventType,
	}
}

func fixModelAPISpecInputWithFetchRequest() *model.SpecInput {
	var specData = "specData"
	var apiType = model.APISpecTypeOdata
	return &model.SpecInput{
		Data: &specData,
		FetchRequest: &model.FetchRequestInput{
			URL: "foo.bar",
		},
		APIType: &apiType,
		Format:  model.SpecFormatXML,
	}
}

func fixModelEventSpecInputWithFetchRequest() *model.SpecInput {
	var specData = "specData"
	var eventType = model.EventSpecTypeAsyncAPI
	return &model.SpecInput{
		Data: &specData,
		FetchRequest: &model.FetchRequestInput{
			URL: "foo.bar",
		},
		EventType: &eventType,
		Format:    model.SpecFormatJSON,
	}
}

func fixGQLAPISpecInput() *graphql.APISpecInput {
	var specData = "specData"
	clob := graphql.CLOB(specData)
	return &graphql.APISpecInput{
		Data:   &clob,
		Type:   graphql.APISpecTypeOdata,
		Format: graphql.SpecFormatXML,
	}
}

func fixGQLAPISpecInputWithFetchRequest() *graphql.APISpecInput {
	var specData = "specData"
	clob := graphql.CLOB(specData)
	return &graphql.APISpecInput{
		Data: &clob,
		FetchRequest: &graphql.FetchRequestInput{
			URL: "foo.bar",
		},
		Type:   graphql.APISpecTypeOdata,
		Format: graphql.SpecFormatXML,
	}
}

func fixGQLEventSpecInput() *graphql.EventSpecInput {
	var specData = "specData"
	clob := graphql.CLOB(specData)
	return &graphql.EventSpecInput{
		Data:   &clob,
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatJSON,
	}
}

func fixGQLEventSpecInputWithFetchRequest() *graphql.EventSpecInput {
	var specData = "specData"
	clob := graphql.CLOB(specData)
	return &graphql.EventSpecInput{
		Data: &clob,
		FetchRequest: &graphql.FetchRequestInput{
			URL: "foo.bar",
		},
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatJSON,
	}
}

func fixSpecColumns() []string {
	return []string{"id", "api_def_id", "event_def_id", "spec_data", "api_spec_format", "api_spec_type", "event_spec_format", "event_spec_type", "custom_type"}
}

func fixAPISpecRow() []driver.Value {
	return []driver.Value{specID, apiID, nil, "specData", "XML", "ODATA", nil, nil, nil}
}

func fixAPISpecRowWithID(id string) []driver.Value {
	return []driver.Value{id, apiID, nil, "specData", "XML", "ODATA", nil, nil, nil}
}

func fixAPISpecRowWithIDs(id, apiDefID string) []driver.Value {
	return []driver.Value{id, apiDefID, nil, "specData", "XML", "ODATA", nil, nil, nil}
}

func fixEventSpecRow() []driver.Value {
	return []driver.Value{specID, nil, eventID, "specData", nil, nil, "JSON", "ASYNC_API", nil}
}

func fixEventSpecRowWithID(id string) []driver.Value {
	return []driver.Value{id, nil, eventID, "specData", nil, nil, "JSON", "ASYNC_API", nil}
}

func fixEventSpecRowWithIDs(id, eventDefID string) []driver.Value {
	return []driver.Value{id, nil, eventDefID, "specData", nil, nil, "JSON", "ASYNC_API", nil}
}

func fixAPISpecCreateArgs(spec *model.Spec) []driver.Value {
	return []driver.Value{specID, spec.ObjectID, nil, spec.Data, spec.DataHash, spec.Format, spec.APIType, nil, spec.EventType, spec.CustomType}
}

func fixEventSpecCreateArgs(spec *model.Spec) []driver.Value {
	return []driver.Value{specID, nil, spec.ObjectID, spec.Data, spec.DataHash, nil, spec.APIType, spec.Format, spec.EventType, spec.CustomType}
}

func fixAPISpecEntity() *spec.Entity {
	return &spec.Entity{
		ID:            specID,
		APIDefID:      repo.NewValidNullableString(apiID),
		SpecData:      repo.NewValidNullableString("specData"),
		APISpecFormat: repo.NewValidNullableString("XML"),
		APISpecType:   repo.NewValidNullableString(string(model.APISpecTypeOdata)),
	}
}

func fixAPISpecEntityWithID(id string) spec.Entity {
	return spec.Entity{
		ID:            id,
		APIDefID:      repo.NewValidNullableString(apiID),
		SpecData:      repo.NewValidNullableString("specData"),
		APISpecFormat: repo.NewValidNullableString("XML"),
		APISpecType:   repo.NewValidNullableString(string(model.APISpecTypeOdata)),
	}
}

func fixAPISpecEntityWithIDs(id, apiDefID string) spec.Entity {
	return spec.Entity{
		ID:            id,
		APIDefID:      repo.NewValidNullableString(apiDefID),
		SpecData:      repo.NewValidNullableString("specData"),
		APISpecFormat: repo.NewValidNullableString("XML"),
		APISpecType:   repo.NewValidNullableString(string(model.APISpecTypeOdata)),
	}
}

func fixEventSpecEntity() *spec.Entity {
	return &spec.Entity{
		ID:              specID,
		EventAPIDefID:   repo.NewValidNullableString(eventID),
		SpecData:        repo.NewValidNullableString("specData"),
		EventSpecType:   repo.NewValidNullableString(string(model.EventSpecTypeAsyncAPI)),
		EventSpecFormat: repo.NewValidNullableString("JSON"),
	}
}

func fixEventSpecEntityWithID(id string) spec.Entity {
	return spec.Entity{
		ID:              id,
		EventAPIDefID:   repo.NewValidNullableString(eventID),
		SpecData:        repo.NewValidNullableString("specData"),
		EventSpecType:   repo.NewValidNullableString(string(model.EventSpecTypeAsyncAPI)),
		EventSpecFormat: repo.NewValidNullableString("JSON"),
	}
}

func fixEventSpecEntityWithIDs(id, eventDefID string) spec.Entity {
	return spec.Entity{
		ID:              id,
		EventAPIDefID:   repo.NewValidNullableString(eventDefID),
		SpecData:        repo.NewValidNullableString("specData"),
		EventSpecType:   repo.NewValidNullableString(string(model.EventSpecTypeAsyncAPI)),
		EventSpecFormat: repo.NewValidNullableString("JSON"),
	}
}
