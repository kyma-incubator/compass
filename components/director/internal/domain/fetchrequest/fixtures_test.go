package fetchrequest_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
