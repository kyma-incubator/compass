package bundleinstanceauth_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

var (
	testID             = "foo"
	testBundleID       = "bar"
	testRuntimeID      = "d05fb90c-3084-4349-9deb-af23a4ce76be"
	testTenant         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	testExternalTenant = "foobaz"
	testContext        = `{"foo": "bar"}`
	testInputParams    = `{"bar": "baz"}`
	testError          = errors.New("test")
	testTime           = time.Now()
	testTableColumns   = []string{"id", "owner_id", "bundle_id", "context", "input_params", "auth_value", "status_condition", "status_timestamp", "status_message", "status_reason", "runtime_id", "runtime_context_id"}
)

func fixModelBundleInstanceAuth(id, bundleID, tenant string, auth *model.Auth, status *model.BundleInstanceAuthStatus, runtimeID *string) *model.BundleInstanceAuth {
	pia := fixModelBundleInstanceAuthWithoutContextAndInputParams(id, bundleID, tenant, auth, status, runtimeID)
	pia.Context = &testContext
	pia.InputParams = &testInputParams

	return pia
}

func fixModelBundleInstanceAuthWithoutContextAndInputParams(id, bundleID, tenant string, auth *model.Auth, status *model.BundleInstanceAuthStatus, runtimeID *string) *model.BundleInstanceAuth {
	return &model.BundleInstanceAuth{
		ID:        id,
		BundleID:  bundleID,
		RuntimeID: runtimeID,
		Owner:     tenant,
		Auth:      auth,
		Status:    status,
	}
}

func fixGQLBundleInstanceAuth(id string, auth *graphql.Auth, status *graphql.BundleInstanceAuthStatus, runtimeID *string) *graphql.BundleInstanceAuth {
	context := graphql.JSON(testContext)
	inputParams := graphql.JSON(testInputParams)

	out := fixGQLBundleInstanceAuthWithoutContextAndInputParams(id, auth, status, runtimeID)
	out.Context = &context
	out.InputParams = &inputParams

	return out
}

func fixGQLBundleInstanceAuthWithoutContextAndInputParams(id string, auth *graphql.Auth, status *graphql.BundleInstanceAuthStatus, runtimeID *string) *graphql.BundleInstanceAuth {
	return &graphql.BundleInstanceAuth{
		ID:        id,
		Auth:      auth,
		Status:    status,
		RuntimeID: runtimeID,
	}
}

func fixModelStatusSucceeded() *model.BundleInstanceAuthStatus {
	return &model.BundleInstanceAuthStatus{
		Condition: model.BundleInstanceAuthStatusConditionSucceeded,
		Timestamp: testTime,
		Message:   "Credentials were provided.",
		Reason:    "CredentialsProvided",
	}
}

func fixModelStatusPending() *model.BundleInstanceAuthStatus {
	return &model.BundleInstanceAuthStatus{
		Condition: model.BundleInstanceAuthStatusConditionPending,
		Timestamp: testTime,
		Message:   "Credentials were not yet provided.",
		Reason:    "CredentialsNotProvided",
	}
}

func fixGQLStatusSucceeded() *graphql.BundleInstanceAuthStatus {
	return &graphql.BundleInstanceAuthStatus{
		Condition: graphql.BundleInstanceAuthStatusConditionSucceeded,
		Timestamp: graphql.Timestamp(testTime),
		Message:   "Credentials were provided.",
		Reason:    "CredentialsProvided",
	}
}

func fixGQLStatusPending() *graphql.BundleInstanceAuthStatus {
	return &graphql.BundleInstanceAuthStatus{
		Condition: graphql.BundleInstanceAuthStatusConditionPending,
		Timestamp: graphql.Timestamp(testTime),
		Message:   "Credentials were not yet provided.",
		Reason:    "CredentialsNotProvided",
	}
}

func fixModelStatusInput(condition model.BundleInstanceAuthSetStatusConditionInput, message, reason string) *model.BundleInstanceAuthStatusInput {
	return &model.BundleInstanceAuthStatusInput{
		Condition: condition,
		Message:   message,
		Reason:    reason,
	}
}

func fixGQLStatusInput(condition graphql.BundleInstanceAuthSetStatusConditionInput, message, reason string) *graphql.BundleInstanceAuthStatusInput {
	return &graphql.BundleInstanceAuthStatusInput{
		Condition: condition,
		Message:   message,
		Reason:    reason,
	}
}

func fixModelRequestInput() *model.BundleInstanceAuthRequestInput {
	return &model.BundleInstanceAuthRequestInput{
		Context:     &testContext,
		InputParams: &testInputParams,
	}
}

func fixModelCreateInput() *model.BundleInstanceAuthCreateInput {
	return &model.BundleInstanceAuthCreateInput{
		Context:     &testContext,
		InputParams: &testInputParams,
		Auth:        fixModelAuthInput(),
		RuntimeID:   &testRuntimeID,
	}
}

func fixModelUpdateInput() *model.BundleInstanceAuthUpdateInput {
	return &model.BundleInstanceAuthUpdateInput{
		Context:     &testContext,
		InputParams: &testInputParams,
		Auth:        fixModelAuthInput(),
	}
}

func fixGQLRequestInput() *graphql.BundleInstanceAuthRequestInput {
	context := graphql.JSON(testContext)
	inputParams := graphql.JSON(testInputParams)

	return &graphql.BundleInstanceAuthRequestInput{
		Context:     &context,
		InputParams: &inputParams,
	}
}

func fixGQLCreateInput() *graphql.BundleInstanceAuthCreateInput {
	context := graphql.JSON(testContext)
	inputParams := graphql.JSON(testInputParams)

	return &graphql.BundleInstanceAuthCreateInput{
		Context:     &context,
		InputParams: &inputParams,
		Auth:        fixGQLAuthInput(),
		RuntimeID:   &testRuntimeID,
	}
}

func fixGQLUpdateInput() *graphql.BundleInstanceAuthUpdateInput {
	context := graphql.JSON(testContext)
	inputParams := graphql.JSON(testInputParams)

	return &graphql.BundleInstanceAuthUpdateInput{
		Context:     &context,
		InputParams: &inputParams,
		Auth:        fixGQLAuthInput(),
	}
}

func fixModelSetInput() *model.BundleInstanceAuthSetInput {
	return &model.BundleInstanceAuthSetInput{
		Auth:   fixModelAuthInput(),
		Status: fixModelStatusInput(model.BundleInstanceAuthSetStatusConditionInputSucceeded, "foo", "bar"),
	}
}

func fixGQLSetInput() *graphql.BundleInstanceAuthSetInput {
	return &graphql.BundleInstanceAuthSetInput{
		Auth:   fixGQLAuthInput(),
		Status: fixGQLStatusInput(graphql.BundleInstanceAuthSetStatusConditionInputSucceeded, "foo", "bar"),
	}
}

func fixEntityBundleInstanceAuth(t *testing.T, id, bundleID, tenant string, auth *model.Auth, status *model.BundleInstanceAuthStatus, runtimeID *string) *bundleinstanceauth.Entity {
	out := fixEntityBundleInstanceAuthWithoutContextAndInputParams(t, id, bundleID, tenant, auth, status, runtimeID)
	out.Context = sql.NullString{Valid: true, String: testContext}
	out.InputParams = sql.NullString{Valid: true, String: testInputParams}

	return out
}

func fixEntityBundleInstanceAuthWithoutContextAndInputParams(t *testing.T, id, bundleID, tenant string, auth *model.Auth, status *model.BundleInstanceAuthStatus, runtimeID *string) *bundleinstanceauth.Entity {
	sqlNullString := sql.NullString{}
	if runtimeID != nil {
		sqlNullString.Valid = true
		sqlNullString.String = *runtimeID
	}
	out := bundleinstanceauth.Entity{
		ID:        id,
		BundleID:  bundleID,
		RuntimeID: sqlNullString,
		OwnerID:   tenant,
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
		out.StatusMessage = status.Message
		out.StatusReason = status.Reason
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
	id               string
	ownerID          string
	bundleID         string
	runtimeID        sql.NullString
	runtimeContextID sql.NullString
	context          sql.NullString
	inputParams      sql.NullString
	authValue        sql.NullString
	statusCondition  string
	statusTimestamp  time.Time
	statusMessage    string
	statusReason     string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.ownerID, row.bundleID, row.context, row.inputParams, row.authValue, row.statusCondition, row.statusTimestamp, row.statusMessage, row.statusReason, row.runtimeID, row.runtimeContextID)
	}
	return out
}

func fixSQLRowFromEntity(entity bundleinstanceauth.Entity) sqlRow {
	return sqlRow{
		id:               entity.ID,
		ownerID:          entity.OwnerID,
		bundleID:         entity.BundleID,
		runtimeID:        entity.RuntimeID,
		runtimeContextID: entity.RuntimeContextID,
		context:          entity.Context,
		inputParams:      entity.InputParams,
		authValue:        entity.AuthValue,
		statusCondition:  entity.StatusCondition,
		statusTimestamp:  entity.StatusTimestamp,
		statusMessage:    entity.StatusMessage,
		statusReason:     entity.StatusReason,
	}
}

func fixCreateArgs(ent bundleinstanceauth.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.OwnerID, ent.BundleID, ent.Context, ent.InputParams, ent.AuthValue, ent.StatusCondition, ent.StatusTimestamp, ent.StatusMessage, ent.StatusReason, ent.RuntimeID, ent.RuntimeContextID}
}

func fixSimpleModelBundleInstanceAuth(id string) *model.BundleInstanceAuth {
	return &model.BundleInstanceAuth{
		ID:       id,
		BundleID: testBundleID,
	}
}

func fixSimpleGQLBundleInstanceAuth(id string) *graphql.BundleInstanceAuth {
	return &graphql.BundleInstanceAuth{
		ID: id,
	}
}

func fixModelBundle(id string, requestInputSchema *string, defaultAuth *model.Auth) *model.Bundle {
	return &model.Bundle{
		ApplicationID:                  "foo",
		Name:                           "test-bundle",
		InstanceAuthRequestInputSchema: requestInputSchema,
		DefaultInstanceAuth:            defaultAuth,
		BaseEntity:                     &model.BaseEntity{ID: id},
	}
}
