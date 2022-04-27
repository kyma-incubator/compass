package integrationsystem_test

import (
	"database/sql/driver"
	"errors"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	testTenant         = "tnt"
	testExternalTenant = "external-tnt"
	testID             = "foo"
	testName           = "bar"
	testPageSize       = 3
	testCursor         = ""
)

var (
	testError        = errors.New("test error")
	testDescription  = "bazz"
	testTableColumns = []string{"id", "name", "description"}
)

func fixModelIntegrationSystem(id, name string) *model.IntegrationSystem {
	return &model.IntegrationSystem{
		ID:          id,
		Name:        name,
		Description: &testDescription,
	}
}

func fixGQLIntegrationSystem(id, name string) *graphql.IntegrationSystem {
	return &graphql.IntegrationSystem{
		ID:          id,
		Name:        name,
		Description: &testDescription,
	}
}

func fixModelIntegrationSystemInput(name string) model.IntegrationSystemInput {
	return model.IntegrationSystemInput{
		Name:        name,
		Description: &testDescription,
	}
}

func fixGQLIntegrationSystemInput(name string) graphql.IntegrationSystemInput {
	return graphql.IntegrationSystemInput{
		Name:        name,
		Description: &testDescription,
	}
}

func fixEntityIntegrationSystem(id, name string) *integrationsystem.Entity {
	return &integrationsystem.Entity{
		ID:          id,
		Name:        name,
		Description: &testDescription,
	}
}

type sqlRow struct {
	id          string
	name        string
	description *string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.description)
	}
	return out
}

func fixIntegrationSystemCreateArgs(ent integrationsystem.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.Description}
}

func fixModelIntegrationSystemPage(intSystems []*model.IntegrationSystem) model.IntegrationSystemPage {
	return model.IntegrationSystemPage{
		Data: intSystems,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(intSystems),
	}
}

func fixGQLIntegrationSystemPage(intSystems []*graphql.IntegrationSystem) graphql.IntegrationSystemPage {
	return graphql.IntegrationSystemPage{
		Data: intSystems,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(intSystems),
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

func fixModelSystemAuth(id, intSysID string, auth *model.Auth) pkgmodel.SystemAuth {
	return pkgmodel.SystemAuth{
		ID:                  id,
		TenantID:            nil,
		IntegrationSystemID: &intSysID,
		Value:               auth,
	}
}

func fixGQLSystemAuth(id string, auth *graphql.Auth) *graphql.IntSysSystemAuth {
	return &graphql.IntSysSystemAuth{
		ID:   id,
		Auth: auth,
	}
}
