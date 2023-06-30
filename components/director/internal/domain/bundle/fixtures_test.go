package bundle_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"

	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

var fixedTimestamp = time.Now()

func fixModelAPIDefinition(id string, name, description string, group string) *model.APIDefinition {
	return &model.APIDefinition{
		Name:        name,
		Description: &description,
		Group:       &group,
		BaseEntity:  &model.BaseEntity{ID: id},
	}
}

func fixGQLAPIDefinition(id string, bndlID string, name, description string, group string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		BundleID:    bndlID,
		Name:        name,
		Description: &description,
		Group:       &group,
	}
}

func fixAPIDefinitionPage(apiDefinitions []*model.APIDefinition) *model.APIDefinitionPage {
	return &model.APIDefinitionPage{
		Data: apiDefinitions,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(apiDefinitions),
	}
}

func fixGQLAPIDefinitionPage(apiDefinitions []*graphql.APIDefinition) *graphql.APIDefinitionPage {
	return &graphql.APIDefinitionPage{
		Data: apiDefinitions,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(apiDefinitions),
	}
}

func fixModelEventAPIDefinition(id string, name, description string, group string) *model.EventDefinition {
	return &model.EventDefinition{
		Name:        name,
		Description: &description,
		Group:       &group,
		BaseEntity:  &model.BaseEntity{ID: id},
	}
}

func fixGQLEventDefinition(id string, bundleID string, name, description string, group string) *graphql.EventDefinition {
	return &graphql.EventDefinition{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		BundleID:    bundleID,
		Name:        name,
		Description: &description,
		Group:       &group,
	}
}

func fixEventAPIDefinitionPage(eventAPIDefinitions []*model.EventDefinition) *model.EventDefinitionPage {
	return &model.EventDefinitionPage{
		Data: eventAPIDefinitions,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(eventAPIDefinitions),
	}
}

func fixGQLEventDefinitionPage(eventAPIDefinitions []*graphql.EventDefinition) *graphql.EventDefinitionPage {
	return &graphql.EventDefinitionPage{
		Data: eventAPIDefinitions,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(eventAPIDefinitions),
	}
}

var (
	docKind              = "fookind"
	docTitle             = "footitle"
	docData              = "foodata"
	docCLOB              = graphql.CLOB(docData)
	desc                 = "Lorem Ipsum"
	appID                = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

func fixModelDocument(bundleID, id string) *model.Document {
	return &model.Document{
		BundleID:   bundleID,
		Title:      docTitle,
		Format:     model.DocumentFormatMarkdown,
		Kind:       &docKind,
		Data:       &docData,
		BaseEntity: &model.BaseEntity{ID: id},
	}
}

func fixModelDocumentPage(documents []*model.Document) *model.DocumentPage {
	return &model.DocumentPage{
		Data: documents,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(documents),
	}
}

func fixGQLDocument(id string) *graphql.Document {
	return &graphql.Document{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		Title:  docTitle,
		Format: graphql.DocumentFormatMarkdown,
		Kind:   &docKind,
		Data:   &docCLOB,
	}
}

func fixGQLDocumentPage(documents []*graphql.Document) *graphql.DocumentPage {
	return &graphql.DocumentPage{
		Data: documents,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(documents),
	}
}

const (
	bundleID             = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	secondBundleID       = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	tenantID             = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID     = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	consumerID           = "consumerID"
	bundleAuthName       = "bundleAuthName"
	bundleAuthID         = "bundleAuthID"
	otherBundleAuthName  = "otherBundleAuthName"
	otherBundleAuthID    = "otherBundleAuthID"
	unusedBundleAuthName = "unusedBundleAuthName"
	unusedBundleAuthID   = "unusedBundleAuthID"
	ordID                = "com.compass.v1"
	localTenantID        = "localTenantID"
	correlationIDs       = `["id1", "id2"]`
	version              = "1.2.2"
	resourceHash         = "123456"
)

func fixBundleModel(name, desc string) *model.Bundle {
	return fixBundleModelWithIDAndAppID(bundleID, name, desc)
}

func fixBundleModelWithID(id, name, desc string) *model.Bundle {
	return fixBundleModelWithAuth(id, name, desc, fixModelAuth())
}

func fixBundleModelWithAuth(id, name, desc string, auth *model.Auth) *model.Bundle {
	return &model.Bundle{
		Name:                           name,
		Description:                    &desc,
		Version:                        str.Ptr(version),
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            auth,
		OrdID:                          str.Ptr(ordID),
		LocalTenantID:                  str.Ptr(localTenantID),
		ShortDescription:               str.Ptr("short_description"),
		Links:                          json.RawMessage("[]"),
		Labels:                         json.RawMessage("[]"),
		CredentialExchangeStrategies:   json.RawMessage("[]"),
		CorrelationIDs:                 json.RawMessage(correlationIDs),
		Tags:                           json.RawMessage("[]"),
		ResourceHash:                   str.Ptr(resourceHash),
		DocumentationLabels:            json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixBundleModelWithIDAndAppID(id, name, desc string) *model.Bundle {
	bundle := fixBundleModelWithID(id, name, desc)
	bundle.ApplicationID = &appID
	return bundle
}
func fixBundleModelWithIDAndAppTemplateVersionID(id, name, desc string) *model.Bundle {
	bundle := fixBundleModelWithID(id, name, desc)
	bundle.ApplicationTemplateVersionID = &appTemplateVersionID
	return bundle
}

func fixGQLBundle(id, name, desc string) *graphql.Bundle {
	schema := graphql.JSONSchema(`{"$id":"https://example.com/person.schema.json","$schema":"http://json-schema.org/draft-07/schema#","properties":{"age":{"description":"Age in years which must be equal to or greater than zero.","minimum":0,"type":"integer"},"firstName":{"description":"The person's first name.","type":"string"},"lastName":{"description":"The person's last name.","type":"string"}},"title":"Person","type":"object"}`)
	var correlationIDsAsSlice []string
	err := json.Unmarshal([]byte(correlationIDs), &correlationIDsAsSlice)
	if err != nil {
		panic(err)
	}
	return &graphql.Bundle{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: &schema,
		DefaultInstanceAuth:            fixGQLAuth(),
		CorrelationIDs:                 correlationIDsAsSlice,
		BaseEntity: &graphql.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
		},
	}
}

func fixGQLBundleCreateInput(name, description string) graphql.BundleCreateInput {
	basicCredentialDataInput := graphql.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}

	credentialDataInput := graphql.CredentialDataInput{Basic: &basicCredentialDataInput}
	defaultAuth := graphql.AuthInput{
		Credential: &credentialDataInput,
	}

	return graphql.BundleCreateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicInputSchema(),
		DefaultInstanceAuth:            &defaultAuth,
		APIDefinitions: []*graphql.APIDefinitionInput{
			{Name: "api1", TargetURL: "foo.bar"},
			{Name: "api2", TargetURL: "foo.bar2"},
		},
		EventDefinitions: []*graphql.EventDefinitionInput{
			{Name: "event1", Description: &desc},
			{Name: "event2", Description: &desc},
		},
		Documents: []*graphql.DocumentInput{
			{DisplayName: "doc1", Kind: &docKind},
			{DisplayName: "doc2", Kind: &docKind},
		},
	}
}

func fixModelBundleCreateInput(name, description string) model.BundleCreateInput {
	basicCredentialDataInput := model.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	authInput := model.AuthInput{
		Credential: &model.CredentialDataInput{Basic: &basicCredentialDataInput},
	}

	specData1 := "spec_data1"
	specData2 := "spec_data2"

	return model.BundleCreateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &authInput,
		APIDefinitions: []*model.APIDefinitionInput{
			{Name: "api1", TargetURLs: api.ConvertTargetURLToJSONArray("foo.bar")},
			{Name: "api2", TargetURLs: api.ConvertTargetURLToJSONArray("foo.bar2")},
		},
		APISpecs: []*model.SpecInput{
			{Data: &specData1},
			{Data: &specData2},
		},
		EventDefinitions: []*model.EventDefinitionInput{
			{Name: "event1", Description: &desc},
			{Name: "event2", Description: &desc},
		},
		EventSpecs: []*model.SpecInput{
			{Data: &specData1},
			{Data: &specData2},
		},
		Documents: []*model.DocumentInput{
			{DisplayName: "doc1", Kind: &docKind},
			{DisplayName: "doc2", Kind: &docKind},
		},
	}
}

func fixGQLBundleUpdateInput(name, description string) graphql.BundleUpdateInput {
	basicCredentialDataInput := graphql.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}

	credentialDataInput := graphql.CredentialDataInput{Basic: &basicCredentialDataInput}
	defaultAuth := graphql.AuthInput{
		Credential: &credentialDataInput,
	}

	return graphql.BundleUpdateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicInputSchema(),
		DefaultInstanceAuth:            &defaultAuth,
	}
}

func fixModelBundleUpdateInput(name, description string) model.BundleUpdateInput {
	basicCredentialDataInput := model.BasicCredentialDataInput{
		Username: "test",
		Password: "pwd",
	}
	authInput := model.AuthInput{
		Credential: &model.CredentialDataInput{Basic: &basicCredentialDataInput},
	}

	return model.BundleUpdateInput{
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &authInput,
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

func fixModelAuthWithUsername(username string) *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: username,
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
		AdditionalHeaders:     graphql.HTTPHeaders{"test": {"foo", "bar"}},
		AdditionalQueryParams: graphql.QueryParams{"test": {"foo", "bar"}},
		RequestAuth: &graphql.CredentialRequestAuth{
			Csrf: &graphql.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: "foo.url",
				Credential: &graphql.BasicCredentialData{
					Username: "boo",
					Password: "far",
				},
				AdditionalHeaders:     graphql.HTTPHeaders{"test": {"foo", "bar"}},
				AdditionalQueryParams: graphql.QueryParams{"test": {"foo", "bar"}},
			},
		},
	}
}

func fixEntityBundleWithAppID(id, name, desc string) *bundle.Entity {
	entity := fixEntityBundle(id, name, desc)
	entity.ApplicationID = repo.NewValidNullableString(appID)
	return entity
}

func fixEntityBundleWithAppTemplateVersionID(id, name, desc string) *bundle.Entity {
	entity := fixEntityBundle(id, name, desc)
	entity.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)
	return entity
}

func fixEntityBundle(id, name, desc string) *bundle.Entity {
	descSQL := sql.NullString{String: desc, Valid: true}
	schemaSQL := sql.NullString{
		String: inputSchemaString(),
		Valid:  true,
	}
	authSQL := sql.NullString{
		String: `{"Credential":{"Basic":{"Username":"foo","Password":"bar"},"Oauth":null,"CertificateOAuth":null},"AccessStrategy":null,"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]},"RequestAuth":{"Csrf":{"TokenEndpointURL":"foo.url","Credential":{"Basic":{"Username":"boo","Password":"far"},"Oauth":null,"CertificateOAuth":null},"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]}}},"OneTimeToken":null,"CertCommonName":""}`,
		Valid:  true,
	}

	return &bundle.Entity{
		Name:                          name,
		Description:                   descSQL,
		Version:                       repo.NewValidNullableString(version),
		InstanceAuthRequestJSONSchema: schemaSQL,
		DefaultInstanceAuth:           authSQL,
		OrdID:                         repo.NewValidNullableString(ordID),
		LocalTenantID:                 repo.NewValidNullableString(localTenantID),
		ShortDescription:              repo.NewValidNullableString("short_description"),
		Links:                         repo.NewValidNullableString("[]"),
		Labels:                        repo.NewValidNullableString("[]"),
		CredentialExchangeStrategies:  repo.NewValidNullableString("[]"),
		CorrelationIDs:                repo.NewValidNullableString(correlationIDs),
		Tags:                          repo.NewValidNullableString("[]"),
		ResourceHash:                  repo.NewValidNullableString(resourceHash),
		DocumentationLabels:           repo.NewValidNullableString("[]"),
		BaseEntity: &repo.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     sql.NullString{},
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixBundleColumns() []string {
	return []string{"id", "app_id", "app_template_version_id", "name", "description", "version", "instance_auth_request_json_schema", "default_instance_auth", "ord_id", "local_tenant_id", "short_description", "links", "labels", "credential_exchange_strategies", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "tags", "resource_hash", "documentation_labels"}
}

func fixBundleRowWithAppID(id string) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(""), "foo", "bar", version, fixSchema(), fixDefaultAuth(), ordID, localTenantID, str.Ptr("short_description"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), true, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(correlationIDs), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]")}
}

func fixBundleRowWithAppTemplateVersionID(id string) []driver.Value {
	return []driver.Value{id, repo.NewValidNullableString(""), appTemplateVersionID, "foo", "bar", version, fixSchema(), fixDefaultAuth(), ordID, localTenantID, str.Ptr("short_description"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), true, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(correlationIDs), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]")}
}

func fixBundleRowWithCustomAppID(id, applicationID string) []driver.Value {
	return []driver.Value{id, applicationID, repo.NewValidNullableString(""), "foo", "bar", version, fixSchema(), fixDefaultAuth(), ordID, localTenantID, str.Ptr("short_description"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), true, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(correlationIDs), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]")}
}

func fixBundleCreateArgsForApp(defAuth, schema string, bndl *model.Bundle) []driver.Value {
	return []driver.Value{bundleID, appID, repo.NewValidNullableString(""), bndl.Name, bndl.Description, bndl.Version, schema, defAuth, ordID, bndl.LocalTenantID, bndl.ShortDescription, repo.NewNullableStringFromJSONRawMessage(bndl.Links), repo.NewNullableStringFromJSONRawMessage(bndl.Labels), repo.NewNullableStringFromJSONRawMessage(bndl.CredentialExchangeStrategies), bndl.Ready, bndl.CreatedAt, bndl.UpdatedAt, bndl.DeletedAt, bndl.Error, repo.NewNullableStringFromJSONRawMessage(bndl.CorrelationIDs), repo.NewNullableStringFromJSONRawMessage(bndl.Tags), bndl.ResourceHash, repo.NewNullableStringFromJSONRawMessage(bndl.DocumentationLabels)}
}

func fixBundleCreateArgsForAppTemplateVersion(defAuth, schema string, bndl *model.Bundle) []driver.Value {
	return []driver.Value{bundleID, repo.NewValidNullableString(""), appTemplateVersionID, bndl.Name, bndl.Description, bndl.Version, schema, defAuth, ordID, bndl.LocalTenantID, bndl.ShortDescription, repo.NewNullableStringFromJSONRawMessage(bndl.Links), repo.NewNullableStringFromJSONRawMessage(bndl.Labels), repo.NewNullableStringFromJSONRawMessage(bndl.CredentialExchangeStrategies), bndl.Ready, bndl.CreatedAt, bndl.UpdatedAt, bndl.DeletedAt, bndl.Error, repo.NewNullableStringFromJSONRawMessage(bndl.CorrelationIDs), repo.NewNullableStringFromJSONRawMessage(bndl.Tags), bndl.ResourceHash, repo.NewNullableStringFromJSONRawMessage(bndl.DocumentationLabels)}
}

func fixDefaultAuth() string {
	return `{"Credential":{"Basic":{"Username":"foo","Password":"bar"},"Oauth":null,"CertificateOAuth":null},"AccessStrategy":null,"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]},"RequestAuth":{"Csrf":{"TokenEndpointURL":"foo.url","Credential":{"Basic":{"Username":"boo","Password":"far"},"Oauth":null,"CertificateOAuth":null},"AdditionalHeaders":{"test":["foo","bar"]},"AdditionalQueryParams":{"test":["foo","bar"]}}},"OneTimeToken":null,"CertCommonName":""}`
}

func inputSchemaString() string {
	return `{"$id":"https://example.com/person.schema.json","$schema":"http://json-schema.org/draft-07/schema#","properties":{"age":{"description":"Age in years which must be equal to or greater than zero.","minimum":0,"type":"integer"},"firstName":{"description":"The person's first name.","type":"string"},"lastName":{"description":"The person's last name.","type":"string"}},"title":"Person","type":"object"}`
}

func fixBasicInputSchema() *graphql.JSONSchema {
	sch := inputSchemaString()
	jsonSchema := graphql.JSONSchema(sch)
	return &jsonSchema
}

func fixBasicSchema() *string {
	sch := inputSchemaString()
	return &sch
}

func fixSchema() string {
	return `{"$id":"https://example.com/person.schema.json","$schema":"http://json-schema.org/draft-07/schema#","properties":{"age":{"description":"Age in years which must be equal to or greater than zero.","minimum":0,"type":"integer"},"firstName":{"description":"The person's first name.","type":"string"},"lastName":{"description":"The person's last name.","type":"string"}},"title":"Person","type":"object"}`
}

func fixModelBundleInstanceAuth(id string) *model.BundleInstanceAuth {
	status := model.BundleInstanceAuthStatus{
		Condition: model.BundleInstanceAuthStatusConditionPending,
		Timestamp: time.Time{},
		Message:   "test-message",
		Reason:    "test-reason",
	}

	context := "ctx"
	params := "test-param"
	return &model.BundleInstanceAuth{
		ID:          id,
		BundleID:    bundleID,
		Context:     &context,
		InputParams: &params,
		Auth:        fixModelAuth(),
		Status:      &status,
	}
}

func fixModelBundleInstanceAuthWithCustomAuth(id, bundleID string, auth *model.Auth) *model.BundleInstanceAuth {
	status := model.BundleInstanceAuthStatus{
		Condition: model.BundleInstanceAuthStatusConditionPending,
		Timestamp: time.Time{},
		Message:   "test-message",
		Reason:    "test-reason",
	}

	context := "ctx"
	params := "test-param"
	return &model.BundleInstanceAuth{
		ID:          id,
		BundleID:    bundleID,
		Context:     &context,
		InputParams: &params,
		Auth:        auth,
		Status:      &status,
	}
}

func fixGQLBundleInstanceAuth(id string) *graphql.BundleInstanceAuth {
	msg := "test-message"
	reason := "test-reason"
	status := graphql.BundleInstanceAuthStatus{
		Condition: graphql.BundleInstanceAuthStatusConditionPending,
		Timestamp: graphql.Timestamp{},
		Message:   msg,
		Reason:    reason,
	}

	params := graphql.JSON("test-param")
	ctx := graphql.JSON("ctx")
	return &graphql.BundleInstanceAuth{
		ID:          id,
		Context:     &ctx,
		InputParams: &params,
		Auth:        fixGQLAuth(),
		Status:      &status,
	}
}

func fixModelAPIBundleReference(bundleID, apiID string) *model.BundleReference {
	return &model.BundleReference{
		BundleID:            str.Ptr(bundleID),
		ObjectType:          model.BundleAPIReference,
		ObjectID:            str.Ptr(apiID),
		APIDefaultTargetURL: str.Ptr(fmt.Sprintf("https://%s.com", apiID)),
	}
}

func fixModelEventBundleReference(bundleID, eventID string) *model.BundleReference {
	return &model.BundleReference{
		BundleID:   str.Ptr(bundleID),
		ObjectType: model.BundleEventReference,
		ObjectID:   str.Ptr(eventID),
	}
}

func fixBasicModelBundle(id, name string) *model.Bundle {
	return &model.Bundle{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.Auth{},
		BaseEntity: &model.BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}

func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}

// UnusedBundleRepository returns BundleRepository mock that does not expect to be invoked
func UnusedBundleRepository() *automock.BundleRepository {
	return &automock.BundleRepository{}
}

// UnusedBundleInstanceAuthService returns BundleInstanceAuthService mock that does not expect to be invoked
func UnusedBundleInstanceAuthService() *automock.BundleInstanceAuthService {
	return &automock.BundleInstanceAuthService{}
}
