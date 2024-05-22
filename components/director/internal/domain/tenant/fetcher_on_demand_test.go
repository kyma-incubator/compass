package tenant_test

import (
	"context"
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
	var (
		fetchTenantURL    = "https://compass-tenant-fetcher.kyma.local/tenants/v1/fetch"
		tenantID          = "b91b59f7-2563-40b2-aba9-fef726037aa3"
		tenantIDWithSlash = "/path/tenant"
		parentTenantID    = "8d4842ed-0307-4808-85d5-6bbed114c4ff"
		testErr           = errors.New("error")
		ignoreFetchErr    = errors.New("ignore fetching of tenant")
	)

	testCases := []struct {
		Name             string
		TenantID         string
		ParentTenantID   string
		Client           func() *automock.Client
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success when parent ID is present",
			TenantID:       tenantID,
			ParentTenantID: parentTenantID,
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil).Once()
				return client
			},
		},
		{
			Name:           "Success when parent ID is missing",
			TenantID:       tenantID,
			ParentTenantID: "",
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil).Once()
				return client
			},
		},
		{
			Name:           "Error when tenant id contain slash",
			TenantID:       tenantIDWithSlash,
			ParentTenantID: "",
			Client: func() *automock.Client {
				client := &automock.Client{}
				return client
			},
			ExpectedErrorMsg: ignoreFetchErr.Error(),
		},
		{
			Name:           "Error when cannot make the request",
			TenantID:       tenantID,
			ParentTenantID: parentTenantID,
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(nil, testErr).Once()
				return client
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:           "Error when status code is not 200",
			TenantID:       tenantID,
			ParentTenantID: parentTenantID,
			Client: func() *automock.Client {
				client := &automock.Client{}
				client.On("Do", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil).Once()
				return client
			},
			ExpectedErrorMsg: fmt.Sprintf("received status code %d when trying to fetch tenant with ID %s", http.StatusInternalServerError, tenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			httpClient := testCase.Client()
			config := tenant.FetchOnDemandAPIConfig{
				TenantOnDemandURL: fetchTenantURL,
				IsDisabled:        false,
			}
			svc := tenant.NewFetchOnDemandService(httpClient, config)

			// WHEN
			err := svc.FetchOnDemand(context.TODO(), testCase.TenantID, testCase.ParentTenantID)

			// THEN
			if len(testCase.ExpectedErrorMsg) > 0 {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
