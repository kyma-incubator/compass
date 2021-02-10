package fetchrequest_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func fixModelFetchRequest(t *testing.T, url, filter string) *model.FetchRequest {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &model.FetchRequest{
		URL:    url,
		Auth:   &model.Auth{},
		Mode:   model.FetchModeSingle,
		Filter: &filter,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: time,
		},
	}
}

func fixGQLFetchRequest(t *testing.T, url, filter string) *graphql.FetchRequest {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &graphql.FetchRequest{
		URL:    url,
		Auth:   &graphql.Auth{},
		Mode:   graphql.FetchModeSingle,
		Filter: &filter,
		Status: &graphql.FetchRequestStatus{
			Condition: graphql.FetchRequestStatusConditionInitial,
			Timestamp: graphql.Timestamp(time),
		},
	}
}

func fixModelFetchRequestInput(url, filter string) *model.FetchRequestInput {
	mode := model.FetchModeSingle

	return &model.FetchRequestInput{
		URL:    url,
		Auth:   &model.AuthInput{},
		Mode:   &mode,
		Filter: &filter,
	}
}

func fixGQLFetchRequestInput(url, filter string) *graphql.FetchRequestInput {
	mode := graphql.FetchModeSingle

	return &graphql.FetchRequestInput{
		URL:    url,
		Auth:   &graphql.AuthInput{},
		Mode:   &mode,
		Filter: &filter,
	}
}

func fixFullFetchRequestModel(id string, timestamp time.Time) model.FetchRequest {
	filter := "filter"
	return model.FetchRequest{
		ID:     id,
		Tenant: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		URL:    "foo.bar",
		Mode:   model.FetchModeIndex,
		Filter: &filter,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionSucceeded,
			Timestamp: timestamp,
		},
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: "foo",
					Password: "bar",
				},
			},
		},
		ObjectType: model.DocumentFetchRequestReference,
		ObjectID:   "documentID",
	}
}

func fixFullFetchRequestEntity(t *testing.T, id string, timestamp time.Time) fetchrequest.Entity {
	auth := &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
	}

	bytes, err := json.Marshal(auth)
	require.NoError(t, err)

	filter := "filter"
	return fetchrequest.Entity{
		ID:       id,
		TenantID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		URL:      "foo.bar",
		Mode:     string(model.FetchModeIndex),
		Filter: sql.NullString{
			String: filter,
			Valid:  true,
		},
		StatusCondition: string(model.FetchRequestStatusConditionSucceeded),
		StatusTimestamp: timestamp,
		Auth: sql.NullString{
			Valid:  true,
			String: string(bytes),
		},
		SpecID: sql.NullString{},
		DocumentID: sql.NullString{
			Valid:  true,
			String: "documentID",
		},
	}
}

func fixFetchRequestModelWithReference(id string, timestamp time.Time, objectType model.FetchRequestReferenceObjectType, objectID string) model.FetchRequest {
	filter := "filter"
	return model.FetchRequest{
		ID:     id,
		Tenant: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		URL:    "foo.bar",
		Mode:   model.FetchModeIndex,
		Filter: &filter,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionSucceeded,
			Timestamp: timestamp,
		},
		Auth:       nil,
		ObjectType: objectType,
		ObjectID:   objectID,
	}
}

func fixFetchRequestEntityWithReferences(id string, timestamp time.Time, specID, documentID sql.NullString) fetchrequest.Entity {
	return fetchrequest.Entity{
		ID:       id,
		TenantID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		URL:      "foo.bar",
		Mode:     string(model.FetchModeIndex),
		Filter: sql.NullString{
			String: "filter",
			Valid:  true,
		},
		StatusCondition: string(model.FetchRequestStatusConditionSucceeded),
		StatusTimestamp: timestamp,
		Auth:            sql.NullString{},
		SpecID:          specID,
		DocumentID:      documentID,
	}
}
