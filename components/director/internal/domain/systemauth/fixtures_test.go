package systemauth_test

import (
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/DATA-DOG/go-sqlmock"
)

var (
	testTenant           = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	testExternalTenant   = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	testMarshalledSchema = "{\"Credential\":{\"Basic\":{\"Username\":\"foo\",\"Password\":\"bar\"},\"Oauth\":null},\"AccessStrategy\":null,\"AdditionalHeaders\":{\"test\":[\"foo\",\"bar\"]},\"AdditionalQueryParams\":{\"test\":[\"foo\",\"bar\"]},\"RequestAuth\":{\"Csrf\":{\"TokenEndpointURL\":\"foo.url\",\"Credential\":{\"Basic\":{\"Username\":\"boo\",\"Password\":\"far\"},\"Oauth\":null},\"AdditionalHeaders\":{\"test\":[\"foo\",\"bar\"]},\"AdditionalQueryParams\":{\"test\":[\"foo\",\"bar\"]}}},\"OneTimeToken\":null,\"CertCommonName\":\"\"}"
	testErr              = errors.New("test error")
)

var testTableColumns = []string{"id", "tenant_id", "app_id", "runtime_id", "integration_system_id", "value"}

func fixGQLAppSystemAuth(id string, auth *graphql.Auth, refObjID string) graphql.SystemAuth {
	authType := graphql.SystemAuthReferenceTypeApplication
	return &graphql.AppSystemAuth{
		ID:                id,
		Auth:              auth,
		Type:              &authType,
		TenantID:          &testTenant,
		ReferenceObjectID: &refObjID,
	}
}

func fixGQLIntSysSystemAuth(id string, auth *graphql.Auth, refObjID string) graphql.SystemAuth {
	authType := graphql.SystemAuthReferenceTypeIntegrationSystem
	return &graphql.IntSysSystemAuth{
		ID:                id,
		Auth:              auth,
		Type:              &authType,
		TenantID:          nil,
		ReferenceObjectID: &refObjID,
	}
}

func fixGQLRuntimeSystemAuth(id string, auth *graphql.Auth, refObjID string) graphql.SystemAuth {
	authType := graphql.SystemAuthReferenceTypeRuntime
	return &graphql.RuntimeSystemAuth{
		ID:                id,
		Auth:              auth,
		Type:              &authType,
		TenantID:          &testTenant,
		ReferenceObjectID: &refObjID,
	}
}

func fixModelSystemAuth(id string, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string, auth *model.Auth) *pkgmodel.SystemAuth {
	systemAuth := pkgmodel.SystemAuth{
		ID:    id,
		Value: auth,
	}

	switch objectType {
	case pkgmodel.ApplicationReference:
		systemAuth.AppID = &objectID
		systemAuth.TenantID = &testTenant
	case pkgmodel.RuntimeReference:
		systemAuth.RuntimeID = &objectID
		systemAuth.TenantID = &testTenant
	case pkgmodel.IntegrationSystemReference:
		systemAuth.IntegrationSystemID = &objectID
		systemAuth.TenantID = nil
	}

	return &systemAuth
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
		CertCommonName: "",
	}
}

func fixGQLAuthInput() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "foo",
				Password: "bar",
			},
		},
		AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
		AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
		RequestAuth: &graphql.CredentialRequestAuthInput{
			Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: "foo.url",
				Credential: &graphql.CredentialDataInput{
					Basic: &graphql.BasicCredentialDataInput{
						Username: "foo",
						Password: "bar",
					},
				},
				AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
				AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
			},
		},
		CertCommonName: nil,
	}
}

func fixEntity(id string, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string, withAuth bool) systemauth.Entity {
	out := systemauth.Entity{
		ID: id,
	}

	switch objectType {
	case pkgmodel.ApplicationReference:
		out.AppID = repo.NewNullableString(&objectID)
		out.TenantID = repo.NewNullableString(&testTenant)
	case pkgmodel.RuntimeReference:
		out.RuntimeID = repo.NewNullableString(&objectID)
		out.TenantID = repo.NewNullableString(&testTenant)
	case pkgmodel.IntegrationSystemReference:
		out.IntegrationSystemID = repo.NewNullableString(&objectID)
		out.TenantID = repo.NewNullableString(nil)
	}

	if withAuth {
		out.Value = repo.NewNullableString(&testMarshalledSchema)
	}

	return out
}

type sqlRow struct {
	id       string
	tenant   *string
	appID    *string
	rtmID    *string
	intSysID *string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.tenant, row.appID, row.rtmID, row.intSysID, testMarshalledSchema)
	}
	return out
}

func fixSystemAuthCreateArgs(ent systemauth.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.TenantID, ent.AppID, ent.RuntimeID, ent.IntegrationSystemID, ent.Value}
}
