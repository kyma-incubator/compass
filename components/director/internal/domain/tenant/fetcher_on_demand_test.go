package tenant_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchOnDemand(t *testing.T) {
	fetchTenantURL := "https://compass-tenant-fetcher.kyma.local/tenants/v1/fetch"
	tenantID := "b91b59f7-2563-40b2-aba9-fef726037aa3"
	testErr := errors.New("error")

	testCases := []struct {
		Name          string
		Client        func() *automock.Client
		ExpectedError error
	}{
		{
			Name: "Success",
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil).Once()
				return client
			},
		},
		{
			Name: "Error when cannot make the request",
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(nil, testErr).Once()
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name: "Error when status code is not 200",
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil).Once()
				return client
			},
			ExpectedError: errors.New(fmt.Sprintf("received status code %d when trying to fetch tenant with ID %s", http.StatusInternalServerError, tenantID)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			httpClient := testCase.Client()
			svc := tenant.NewFetchOnDemandService(httpClient, fetchTenantURL)

			// WHEN
			err := svc.FetchOnDemand(tenantID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
