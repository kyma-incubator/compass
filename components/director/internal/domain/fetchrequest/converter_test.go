package fetchrequest_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.FetchRequest
		Expected *graphql.FetchRequest
	}{
		{
			Name:     "All properties given",
			Input:    fixModelFetchRequest(t, "foo", "bar"),
			Expected: fixGQLFetchRequest(t, "foo", "bar"),
		},
		{
			Name:  "Empty",
			Input: &model.FetchRequest{},
			Expected: &graphql.FetchRequest{
				Status: &graphql.FetchRequestStatus{
					Condition: graphql.FetchRequestStatusConditionInitial,
				},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("ToGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := fetchrequest.NewConverter(authConv)

			// when
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.FetchRequestInput
		Expected *model.FetchRequestInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLFetchRequestInput("foo", "bar"),
			Expected: fixModelFetchRequestInput("foo", "bar"),
		},
		{
			Name:     "Empty",
			Input:    &graphql.FetchRequestInput{},
			Expected: &model.FetchRequestInput{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("InputFromGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := fetchrequest.NewConverter(authConv)

			// when
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	timestamp := time.Now()

	// given
	testCases := []struct {
		Name               string
		Input              fetchrequest.Entity
		Expected           model.FetchRequest
		ExpectedErrMessage string
	}{
		{
			Name:               "All properties given",
			Input:              fixFetchRequestEntity(t, "1", timestamp),
			Expected:           fixFetchRequestModel("1", timestamp),
			ExpectedErrMessage: "",
		},
		{
			Name: "Empty value",
			Input: fetchrequest.Entity{
				ID:              "2",
				TenantID:        "tenant",
				Auth:            sql.NullString{},
				StatusTimestamp: timestamp,
				StatusCondition: string(model.FetchRequestStatusConditionFailed),
			},
			Expected: model.FetchRequest{
				ID:     "2",
				Tenant: "tenant",
				Auth:   nil,
				Status: &model.FetchRequestStatus{
					Timestamp: timestamp,
					Condition: model.FetchRequestStatusConditionFailed,
				},
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error",
			Input: fetchrequest.Entity{
				Auth: sql.NullString{
					String: `{Dd`,
					Valid:  true,
				},
			},
			Expected:           model.FetchRequest{},
			ExpectedErrMessage: "while unmarshalling Auth: invalid character 'D' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			conv := fetchrequest.NewConverter(authConv)

			// when
			res, err := conv.FromEntity(testCase.Input)

			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	timestamp := time.Now()

	// given
	testCases := []struct {
		Name               string
		Input              model.FetchRequest
		Expected           fetchrequest.Entity
		ExpectedErrMessage string
	}{
		{
			Name:     "All properties given",
			Input:    fixFetchRequestModel("1", timestamp),
			Expected: fixFetchRequestEntity(t, "1", timestamp),
		},
		{
			Name:     "String value",
			Input:    fixFetchRequestModel("1", timestamp),
			Expected: fixFetchRequestEntity(t, "1", timestamp),
		},
		{
			Name: "Empty Auth",
			Input: model.FetchRequest{
				ID:     "2",
				Tenant: "tenant",
				Status: &model.FetchRequestStatus{
					Timestamp: timestamp,
					Condition: model.FetchRequestStatusConditionFailed,
				},
			},
			Expected: fetchrequest.Entity{
				ID:              "2",
				TenantID:        "tenant",
				StatusTimestamp: timestamp,
				StatusCondition: string(model.FetchRequestStatusConditionFailed),
			},
		},
		{
			Name: "Error",
			Input: model.FetchRequest{
				ID:     "2",
				Tenant: "tenant",
			},
			Expected: fetchrequest.Entity{
				ID:       "2",
				TenantID: "tenant",
			},
			ExpectedErrMessage: "Invalid input model",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			conv := fetchrequest.NewConverter(authConv)

			// when
			res, err := conv.ToEntity(testCase.Input)

			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
