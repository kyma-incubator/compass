package apiruntimeauth_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	testTenant           = "tenant"
	testMarshalledSchema = "{\"Credential\":{\"Basic\":{\"Username\":\"foo\",\"Password\":\"bar\"},\"Oauth\":null},\"AdditionalHeaders\":{\"test\":[\"foo\",\"bar\"]},\"AdditionalQueryParams\":{\"test\":[\"foo\",\"bar\"]},\"RequestAuth\":{\"Csrf\":{\"TokenEndpointURL\":\"foo.url\",\"Credential\":{\"Basic\":{\"Username\":\"boo\",\"Password\":\"far\"},\"Oauth\":null},\"AdditionalHeaders\":{\"test\":[\"foo\",\"bar\"]},\"AdditionalQueryParams\":{\"test\":[\"foo\",\"bar\"]}}}}"
)

var testTableColumns = []string{"id", "tenant_id", "runtime_id", "api_def_id", "value"}

func fixGQLAPIRuntimeAuth(runtimeID string, auth *graphql.Auth) *graphql.APIRuntimeAuth {
	return &graphql.APIRuntimeAuth{
		RuntimeID: runtimeID,
		Auth:      auth,
	}
}

func fixModelAPIRuntimeAuth(id *string, runtimeID string, apiID string, auth *model.Auth) *model.APIRuntimeAuth {
	return &model.APIRuntimeAuth{
		ID:        id,
		TenantID:  testTenant,
		RuntimeID: runtimeID,
		APIDefID:  apiID,
		Value:     auth,
	}
}

func fixModelAuthInput() model.AuthInput {
	return model.AuthInput{
		Credential: &model.CredentialDataInput{
			Basic: &model.BasicCredentialDataInput{
				Username: "foo",
				Password: "bar",
			},
		},
		AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
		AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
		RequestAuth: &model.CredentialRequestAuthInput{
			Csrf: &model.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: "foo.url",
				Credential: &model.CredentialDataInput{
					Basic: &model.BasicCredentialDataInput{
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

func fixEntity(id *string, rtmID string, apiID string, withAuth bool) apiruntimeauth.Entity {
	out := apiruntimeauth.Entity{
		TenantID:  testTenant,
		RuntimeID: rtmID,
		APIDefID:  apiID,
	}

	if id != nil {
		out.ID.Valid = true
		out.ID.String = *id
	}
	if withAuth {
		out.Value.Valid = true
		out.Value.String = testMarshalledSchema
	}

	return out
}

type sqlRow struct {
	id    string
	rtmID string
	apiID string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, testTenant, row.rtmID, row.apiID, testMarshalledSchema)
	}
	return out
}
