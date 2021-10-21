package systemauth_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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

func fixGQLAppSystemAuth(id string, auth *graphql.Auth) graphql.SystemAuth {
	return &graphql.AppSystemAuth{
		ID:   id,
		Auth: auth,
	}
}

func fixGQLIntSysSystemAuth(id string, auth *graphql.Auth) graphql.SystemAuth {
	return &graphql.IntSysSystemAuth{
		ID:   id,
		Auth: auth,
	}
}

func fixGQLRuntimeSystemAuth(id string, auth *graphql.Auth) graphql.SystemAuth {
	return &graphql.RuntimeSystemAuth{
		ID:   id,
		Auth: auth,
	}
}

func fixModelSystemAuth(id string, objectType model.SystemAuthReferenceObjectType, objectID string, auth *model.Auth) *model.SystemAuth {
	systemAuth := model.SystemAuth{
		ID:    id,
		Value: auth,
	}

	switch objectType {
	case model.ApplicationReference:
		systemAuth.AppID = &objectID
		systemAuth.TenantID = &testTenant
	case model.RuntimeReference:
		systemAuth.RuntimeID = &objectID
		systemAuth.TenantID = &testTenant
	case model.IntegrationSystemReference:
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

func fixEntity(id string, objectType model.SystemAuthReferenceObjectType, objectID string, withAuth bool) systemauth.Entity {
	out := systemauth.Entity{
		ID: id,
	}

	switch objectType {
	case model.ApplicationReference:
		out.AppID = repo.NewNullableString(&objectID)
		out.TenantID = repo.NewNullableString(&testTenant)
	case model.RuntimeReference:
		out.RuntimeID = repo.NewNullableString(&objectID)
		out.TenantID = repo.NewNullableString(&testTenant)
	case model.IntegrationSystemReference:
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

func fixUpdateTenantIsolationSubquery() string {
	return `tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixUnescapedTenantIsolationSubquery() string {
	return fixUnescapedTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return regexp.QuoteMeta(fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i))
}

func fixUnescapedTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i)
}
