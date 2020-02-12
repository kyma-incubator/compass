package packageinstanceauth_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

var (
	testID           = "foo"
	testPackageID    = "bar"
	testTenant       = "baz"
	testContext      = `{"foo": "bar"}`
	testInputParams  = `{"bar": "baz"}`
	testError        = errors.New("test")
	testTime         = time.Now()
	testTableColumns = []string{"id", "tenant_id", "package_id", "context", "input_params", "auth_value", "status_condition", "status_timestamp", "status_message", "status_reason"}
)

func fixModelPackageInstanceAuth(id, packageID, tenant string, auth *model.Auth, status *model.PackageInstanceAuthStatus) *model.PackageInstanceAuth {
	pia := fixModelPackageInstanceAuthWithoutContextAndInputParams(id, packageID, tenant, auth, status)
	pia.Context = &testContext
	pia.InputParams = &testInputParams

	return pia
}
func fixModelPackageInstanceAuthWithoutContextAndInputParams(id, packageID, tenant string, auth *model.Auth, status *model.PackageInstanceAuthStatus) *model.PackageInstanceAuth {
	return &model.PackageInstanceAuth{
		ID:        id,
		PackageID: packageID,
		Tenant:    tenant,
		Auth:      auth,
		Status:    status,
	}
}

func fixGQLPackageInstanceAuth(id string, auth *graphql.Auth, status *graphql.PackageInstanceAuthStatus) *graphql.PackageInstanceAuth {
	context := graphql.JSON(testContext)
	inputParams := graphql.JSON(testInputParams)

	out := fixGQLPackageInstanceAuthWithoutContextAndInputParams(id, auth, status)
	out.Context = &context
	out.InputParams = &inputParams

	return out
}

func fixGQLPackageInstanceAuthWithoutContextAndInputParams(id string, auth *graphql.Auth, status *graphql.PackageInstanceAuthStatus) *graphql.PackageInstanceAuth {
	return &graphql.PackageInstanceAuth{
		ID:     id,
		Auth:   auth,
		Status: status,
	}
}

func fixModelStatusSucceeded() *model.PackageInstanceAuthStatus {
	return &model.PackageInstanceAuthStatus{
		Condition: model.PackageInstanceAuthStatusConditionSucceeded,
		Timestamp: testTime,
		Message:   str.Ptr("Credentials were provided."),
		Reason:    str.Ptr("CredentialsProvided"),
	}
}

func fixModelStatusPending() *model.PackageInstanceAuthStatus {
	return &model.PackageInstanceAuthStatus{
		Condition: model.PackageInstanceAuthStatusConditionPending,
		Timestamp: testTime,
		Message:   str.Ptr("Credentials were not yet provided."),
		Reason:    str.Ptr("CredentialsNotProvided"),
	}
}

func fixGQLStatusSucceeded() *graphql.PackageInstanceAuthStatus {
	return &graphql.PackageInstanceAuthStatus{
		Condition: graphql.PackageInstanceAuthStatusConditionSucceeded,
		Timestamp: graphql.Timestamp(testTime),
		Message:   "Credentials were provided.",
		Reason:    "CredentialsProvided",
	}
}

func fixGQLStatusPending() *graphql.PackageInstanceAuthStatus {
	return &graphql.PackageInstanceAuthStatus{
		Condition: graphql.PackageInstanceAuthStatusConditionPending,
		Timestamp: graphql.Timestamp(testTime),
		Message:   "Credentials were not yet provided.",
		Reason:    "CredentialsNotProvided",
	}
}

func fixModelStatusInput(condition model.PackageInstanceAuthSetStatusConditionInput, message, reason *string) *model.PackageInstanceAuthStatusInput {
	return &model.PackageInstanceAuthStatusInput{
		Condition: condition,
		Message:   message,
		Reason:    reason,
	}
}

func fixGQLStatusInput(condition graphql.PackageInstanceAuthSetStatusConditionInput, message, reason *string) *graphql.PackageInstanceAuthStatusInput {
	return &graphql.PackageInstanceAuthStatusInput{
		Condition: condition,
		Message:   message,
		Reason:    reason,
	}
}

func fixEntityPackageInstanceAuth(t *testing.T, id, packageID, tenant string, auth *model.Auth, status *model.PackageInstanceAuthStatus) *packageinstanceauth.Entity {
	out := fixEntityPackageInstanceAuthWithoutContextAndInputParams(t, id, packageID, tenant, auth, status)
	out.Context = sql.NullString{Valid: true, String: testContext}
	out.InputParams = sql.NullString{Valid: true, String: testInputParams}

	return out
}

func fixEntityPackageInstanceAuthWithoutContextAndInputParams(t *testing.T, id, packageID, tenant string, auth *model.Auth, status *model.PackageInstanceAuthStatus) *packageinstanceauth.Entity {
	out := packageinstanceauth.Entity{
		ID:        id,
		PackageID: packageID,
		TenantID:  tenant,
	}

	if auth != nil {
		marshalled, err := json.Marshal(auth)
		require.NoError(t, err)
		out.AuthValue = sql.NullString{
			String: string(marshalled),
			Valid:  true,
		}
	}

	if status != nil {
		out.StatusCondition = string(status.Condition)
		out.StatusTimestamp = status.Timestamp

		if status.Message != nil {
			out.StatusMessage = *status.Message
		}

		if status.Reason != nil {
			out.StatusReason = *status.Reason
		}
	}

	return &out
}

func fixModelAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "foo",
				Password: "bar",
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
	}
}

func fixModelAuthInput() *model.AuthInput {
	return &model.AuthInput{
		Credential: &model.CredentialDataInput{
			Basic: &model.BasicCredentialDataInput{
				Username: "foo",
				Password: "bar",
			},
		},
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
	}
}

type sqlRow struct {
	id              string
	tenantID        string
	packageID       string
	context         sql.NullString
	inputParams     sql.NullString
	authValue       sql.NullString
	statusCondition string
	statusTimestamp time.Time
	statusMessage   string
	statusReason    string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.tenantID, row.packageID, row.context, row.inputParams, row.authValue, row.statusCondition, row.statusTimestamp, row.statusMessage, row.statusReason)
	}
	return out
}

func fixSQLRowFromEntity(entity packageinstanceauth.Entity) sqlRow {
	return sqlRow{
		id:              entity.ID,
		tenantID:        entity.TenantID,
		packageID:       entity.PackageID,
		context:         entity.Context,
		inputParams:     entity.InputParams,
		authValue:       entity.AuthValue,
		statusCondition: entity.StatusCondition,
		statusTimestamp: entity.StatusTimestamp,
		statusMessage:   entity.StatusMessage,
		statusReason:    entity.StatusReason,
	}
}

func fixCreateArgs(ent packageinstanceauth.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.TenantID, ent.PackageID, ent.Context, ent.InputParams, ent.AuthValue, ent.StatusCondition, ent.StatusTimestamp, ent.StatusMessage, ent.StatusReason}
}
