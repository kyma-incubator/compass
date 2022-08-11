package resync_test

import (
	"context"
	"fmt"
	"testing"

	domaintenant "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	centralRegion = "central"
	region        = "europe-east"
	provider      = "external-service"
)

func TestTenantManager_TenantsToCreate(t *testing.T) {
	const (
		centralRegion = "central"
		unknownRegion = "europe-east"
		provider      = "external-service"
		timestamp     = "1234567899987"
	)

	ctx := context.TODO()

	// GIVEN
	tenantConverter := domaintenant.NewConverter()

	jobConfig := configForTenantType(tenant.Account)

	busTenant1 := fixBusinessTenantMappingInput("1", provider, "subdomain-1", "", "", tenant.Account)
	busTenant2 := fixBusinessTenantMappingInput("2", provider, "subdomain-2", "", "", tenant.Account)

	event1 := fixEvent(t, "GlobalAccount", busTenant1.ExternalTenant, eventFieldsFromTenant(tenant.Account, jobConfig.APIConfig.TenantFieldMapping, busTenant1))
	event2 := fixEvent(t, "GlobalAccount", busTenant2.ExternalTenant, eventFieldsFromTenant(tenant.Account, jobConfig.APIConfig.TenantFieldMapping, busTenant2))

	pageOneQueryParams := resync.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": timestamp,
	}

	testCases := []struct {
		name              string
		jobConfigFn       func() resync.JobConfig
		region            string
		directorClientFn  func() *automock.DirectorGraphQLClient
		universalClientFn func() *automock.EventAPIClient
		regionalClientsFn func(resync.JobConfig) (map[string]resync.RegionalClient, []*automock.EventAPIClient)
		expectedTenants   []model.BusinessTenantMappingInput
		expectedErrMsg    string
	}{
		{
			name: "Success when only one page is returned from created tenants events",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionalClient, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionalClient)
				details[centralRegion] = resync.RegionalClient{}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Success when two pages are returned from created tenants events",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionalClient, []*automock.EventAPIClient) {
				pageTwoQueryParams := resync.QueryParams{
					"pageSize":  "1",
					"pageNum":   "2",
					"timestamp": timestamp,
				}

				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 2, 2), nil).Once()
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event2), 2, 2), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionalClient)
				details[centralRegion] = resync.RegionalClient{
					RegionalAPIConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:    client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1, busTenant2},
		},
		{
			name: "Success when only one page is returned from updated tenants events",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				return map[string]resync.RegionDetails{
					centralRegion: {
						EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
						RegionalClient:         client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Success when events for both create and update are returned for the same tenant",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(config resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				return map[string]resync.RegionDetails{
					centralRegion: {
						EventAPIRegionalConfig: EventAPIRegionalConfig{
							APIClientConfig:   APIClientConfig{},
							IsUniversalClient: false,
							RegionName:        centralRegion,
						},
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Success when events for both create and update are returned for different tenants",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(config resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event2), 1, 1), nil).Once()
				return map[string]resync.RegionDetails{
					centralRegion: {
						EventAPIRegionalConfig: EventAPIRegionalConfig{
							APIClientConfig:   APIClientConfig{},
							IsUniversalClient: false,
							RegionName:        centralRegion,
						},
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1, busTenant2},
		},
		{
			name: "Success when universal client is enabled",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				reg := cfg.RegionalAPIConfigs[centralRegion]
				reg.IsUniversalClient = true
				cfg.RegionalAPIConfigs[centralRegion] = reg
				return cfg
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "1",
					"timestamp": timestamp,
					"region":    centralRegion,
				}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, queryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, queryParams).Return(nil, nil).Once()
				return client
			},
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Success when regional and universal clients return the same tenant",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				reg := cfg.RegionalAPIConfigs[centralRegion]
				reg.IsUniversalClient = true
				cfg.RegionalAPIConfigs[centralRegion] = reg
				return cfg
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "1",
					"timestamp": timestamp,
					"region":    centralRegion,
				}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, queryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, queryParams).Return(nil, nil).Once()
				return client
			},
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				busTenantWithoutSubdomain := busTenant1
				busTenantWithoutSubdomain.Subdomain = ""
				event := fixEvent(t, "GlobalAccount", busTenant1.ExternalTenant, eventFieldsFromTenant(tenant.Account, cfg.RegionalAPIConfigs[centralRegion].TenantFieldMapping, busTenant1))
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Fail when fetching created tenants events returns an error",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(config resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(nil, errors.New("failed to get created")).Once()
				return map[string]resync.RegionDetails{
					centralRegion: {
						EventAPIRegionalConfig: EventAPIRegionalConfig{
							APIClientConfig:   APIClientConfig{},
							IsUniversalClient: false,
							RegionName:        centralRegion,
						},
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: "while fetching created tenants",
		},
		{
			name: "Fail when fetching updated tenants events returns an error",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(config resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", resync.UpdatedAccountType, pageOneQueryParams).Return(nil, errors.New("failed to get updated")).Once()
				return map[string]resync.RegionDetails{
					centralRegion: {
						EventAPIRegionalConfig: EventAPIRegionalConfig{
							APIClientConfig:   APIClientConfig{},
							IsUniversalClient: false,
							RegionName:        centralRegion,
						},
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: "while fetching updated tenants",
		},
		{
			name: "Fail when universal client returns an error while fetching",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				reg := cfg.RegionalAPIConfigs[centralRegion]
				reg.IsUniversalClient = true
				cfg.RegionalAPIConfigs[centralRegion] = reg
				return cfg
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "1",
					"timestamp": timestamp,
					"region":    centralRegion,
				}
				client.On("FetchTenantEventsPage", resync.CreatedAccountType, queryParams).Return(nil, errors.New("error")).Once()
				return client
			},
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: "while fetching created tenants",
		},
		{
			name: "Fail when region is not supported",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            unknownRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalClientsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: fmt.Sprintf("region %s is not supported", unknownRegion),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.jobConfigFn()

			directorClient := tc.directorClientFn()
			universalClient := tc.universalClientFn()
			regionalDetails, clientMocks := tc.regionalClientsFn(cfg)

			defer func(t *testing.T) {
				mock.AssertExpectationsForObjects(t, directorClient, universalClient)
				for _, clientMock := range clientMocks {
					clientMock.AssertExpectations(t)
				}
			}(t)

			manager, err := resync.NewTenantsManager(cfg, directorClient, universalClient, regionalDetails, tenantConverter)
			require.NoError(t, err)

			res, err := manager.TenantsToCreate(ctx, tc.region, timestamp)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tc.expectedTenants, res)
			}
		})
	}
}

func TestTenantManager_CreateTenants(t *testing.T) {
	const (
		centralRegion = "central"
		region        = "europe-east"
		provider      = "external-service"

		failedToStoreTenantsErrMsg = "failed to store tenants in Director"
	)

	ctx := context.TODO()

	// GIVEN
	tenantConverter := domaintenant.NewConverter()

	busTenant1 := fixBusinessTenantMappingInput("1", provider, "subdomain-1", region, "", tenant.Account)
	busTenant2 := fixBusinessTenantMappingInput("2", provider, "subdomain-2", region, "", tenant.Account)
	busTenants := []model.BusinessTenantMappingInput{busTenant1, busTenant2}

	testCases := []struct {
		name              string
		jobConfigFn       func() resync.JobConfig
		directorClientFn  func() *automock.DirectorGraphQLClient
		universalClientFn func() *automock.EventAPIClient
		regionalDetailsFn func() (map[string]resync.RegionDetails, []*automock.EventAPIClient)
		expectedErrMsg    string
	}{
		{
			name: "Success when tenants are stored in one chunk",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenants))).Return(nil)
				return gqlClient
			},
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func() (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				return map[string]resync.RegionDetails{
					centralRegion: {
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
		},
		{
			name: "Success when tenants are stored in more than one chunk",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				cfg.TenantOperationChunkSize = 1
				return cfg
			},
			directorClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", ctx, tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{busTenant1})).Return(nil).Once()
				gqlClient.On("WriteTenants", ctx, tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{busTenant2})).Return(nil).Once()
				return gqlClient
			},
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func() (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				return map[string]resync.RegionDetails{
					centralRegion: {
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
		},
		{
			name: "Fail when tenant insertion in Director returns an error",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenants))).Return(errors.New(failedToStoreTenantsErrMsg))
				return gqlClient
			},
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func() (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				return map[string]resync.RegionDetails{
					centralRegion: {
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: failedToStoreTenantsErrMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jobCfg := tc.jobConfigFn()
			directorClient := tc.directorClientFn()
			universalClient := tc.universalClientFn()
			regionalDetails, clientMocks := tc.regionalDetailsFn()

			defer func(t *testing.T) {
				mock.AssertExpectationsForObjects(t, directorClient, universalClient)
				for _, clientMock := range clientMocks {
					clientMock.AssertExpectations(t)
				}
			}(t)

			manager, err := NewTenantsManager(jobCfg, directorClient, universalClient, regionalDetails, tenantConverter)
			require.NoError(t, err)

			err = manager.CreateTenants(ctx, busTenants)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenantManager_TenantsToDelete(t *testing.T) {
	const (
		centralRegion = "central"
		unknownRegion = "europe-east"
		provider      = "external-service"
		timestamp     = "1234567899987"
	)

	ctx := context.TODO()

	// GIVEN
	tenantConverter := domaintenant.NewConverter()

	jobConfig := configForTenantType(tenant.Account)

	busTenant1 := fixBusinessTenantMappingInput("1", provider, "subdomain-1", "", "", tenant.Account)
	busTenant2 := fixBusinessTenantMappingInput("2", provider, "subdomain-2", "", "", tenant.Account)

	event1 := fixEvent(t, "GlobalAccount", busTenant1.ExternalTenant, eventFieldsFromTenant(tenant.Account, jobConfig.UniversalClientAPIConfig.TenantFieldMapping, busTenant1))
	event2 := fixEvent(t, "GlobalAccount", busTenant2.ExternalTenant, eventFieldsFromTenant(tenant.Account, jobConfig.UniversalClientAPIConfig.TenantFieldMapping, busTenant2))

	pageOneQueryParams := QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": timestamp,
	}

	testCases := []struct {
		name              string
		jobConfigFn       func() resync.JobConfig
		region            string
		directorClientFn  func() *automock.DirectorGraphQLClient
		universalClientFn func() *automock.EventAPIClient
		regionalDetailsFn func(resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient)
		expectedTenants   []model.BusinessTenantMappingInput
		expectedErrMsg    string
	}{
		{
			name: "Success when only one page is returned from deleted tenants events",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Success when two pages are returned from deleted tenants events",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				pageTwoQueryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "2",
					"timestamp": timestamp,
				}

				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 2, 2), nil).Once()
				client.On("FetchTenantEventsPage", DeletedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event2), 2, 2), nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1, busTenant2},
		},
		{
			name: "Success when universal client is enabled",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				reg := cfg.RegionalAPIConfigs[centralRegion]
				reg.IsUniversalClient = true
				cfg.RegionalAPIConfigs[centralRegion] = reg
				return cfg
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "1",
					"timestamp": timestamp,
					"region":    centralRegion,
				}
				client.On("FetchTenantEventsPage", DeletedAccountType, queryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				return client
			},
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", DeletedAccountType, pageOneQueryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Success when regional and universal clients return the same tenant",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				reg := cfg.RegionalAPIConfigs[centralRegion]
				reg.IsUniversalClient = true
				cfg.RegionalAPIConfigs[centralRegion] = reg
				return cfg
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "1",
					"timestamp": timestamp,
					"region":    centralRegion,
				}
				client.On("FetchTenantEventsPage", DeletedAccountType, queryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				return client
			},
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				busTenantWithoutSubdomain := busTenant1
				busTenantWithoutSubdomain.Subdomain = ""
				event := fixEvent(t, "GlobalAccount", busTenant1.ExternalTenant, eventFieldsFromTenant(tenant.Account, cfg.RegionalAPIConfigs[centralRegion].TenantFieldMapping, busTenant1))
				client.On("FetchTenantEventsPage", DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event), 1, 1), nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenants: []model.BusinessTenantMappingInput{busTenant1},
		},
		{
			name: "Fail when fetching deleted tenants events returns an error",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            centralRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func(config resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", DeletedAccountType, pageOneQueryParams).Return(nil, errors.New("failed to get deleted")).Once()
				return map[string]resync.RegionDetails{
					centralRegion: {
						EventAPIRegionalConfig: EventAPIRegionalConfig{
							APIClientConfig:   APIClientConfig{},
							IsUniversalClient: false,
							RegionName:        centralRegion,
						},
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: "while fetching deleted tenants",
		},
		{
			name: "Fail when universal client returns an error while fetching deleted tenants",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				reg := cfg.RegionalAPIConfigs[centralRegion]
				reg.IsUniversalClient = true
				cfg.RegionalAPIConfigs[centralRegion] = reg
				return cfg
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "1",
					"timestamp": timestamp,
					"region":    centralRegion,
				}
				client.On("FetchTenantEventsPage", DeletedAccountType, queryParams).Return(nil, errors.New("error")).Once()
				return client
			},
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: "while fetching deleted tenants",
		},
		{
			name: "Fail when region is not supported",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			region:            unknownRegion,
			directorClientFn:  func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}

				details := make(map[string]resync.RegionDetails)
				details[centralRegion] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[centralRegion],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: fmt.Sprintf("region %s is not supported", unknownRegion),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.jobConfigFn()

			directorClient := tc.directorClientFn()
			universalClient := tc.universalClientFn()
			regionalDetails, clientMocks := tc.regionalDetailsFn(cfg)

			defer func(t *testing.T) {
				mock.AssertExpectationsForObjects(t, directorClient, universalClient)
				for _, clientMock := range clientMocks {
					clientMock.AssertExpectations(t)
				}
			}(t)

			manager, err := NewTenantsManager(cfg, directorClient, universalClient, regionalDetails, tenantConverter)
			require.NoError(t, err)

			res, err := manager.TenantsToDelete(ctx, tc.region, timestamp)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tc.expectedTenants, res)
			}
		})
	}
}

func TestTenantManager_DeleteTenants(t *testing.T) {
	const failedToDeleteTenantsErrMsg = "failed to delete tenants in Director"

	ctx := context.TODO()

	// GIVEN
	tenantConverter := domaintenant.NewConverter()

	busTenant1 := fixBusinessTenantMappingInput("1", provider, "subdomain-1", region, "", tenant.Account)
	busTenant2 := fixBusinessTenantMappingInput("2", provider, "subdomain-2", region, "", tenant.Account)
	busTenants := []model.BusinessTenantMappingInput{busTenant1, busTenant2}

	testCases := []struct {
		name              string
		jobConfigFn       func() resync.JobConfig
		directorClientFn  func() *automock.DirectorGraphQLClient
		universalClientFn func() *automock.EventAPIClient
		regionalDetailsFn func() (map[string]resync.RegionDetails, []*automock.EventAPIClient)
		expectedErrMsg    string
	}{
		{
			name: "Success when tenants are deleted in one chunk",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", ctx, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenants))).Return(nil)
				return gqlClient
			},
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func() (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				return map[string]resync.RegionDetails{
					centralRegion: {
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
		},
		{
			name: "Success when tenants are deleted in more than one chunk",
			jobConfigFn: func() resync.JobConfig {
				cfg := configForTenantType(tenant.Account)
				cfg.TenantOperationChunkSize = 1
				return cfg
			},
			directorClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", ctx, tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{busTenant1})).Return(nil).Once()
				gqlClient.On("DeleteTenants", ctx, tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{busTenant2})).Return(nil).Once()
				return gqlClient
			},
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func() (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				return map[string]resync.RegionDetails{
					centralRegion: {
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
		},
		{
			name: "Fail when tenant deletion in Director returns an error",
			jobConfigFn: func() resync.JobConfig {
				return configForTenantType(tenant.Account)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", ctx, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenants))).Return(errors.New(failedToDeleteTenantsErrMsg))
				return gqlClient
			},
			universalClientFn: func() *automock.EventAPIClient { return &automock.EventAPIClient{} },
			regionalDetailsFn: func() (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				return map[string]resync.RegionDetails{
					centralRegion: {
						RegionalClient: client,
					},
				}, []*automock.EventAPIClient{client}
			},
			expectedErrMsg: failedToDeleteTenantsErrMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jobCfg := tc.jobConfigFn()
			directorClient := tc.directorClientFn()
			universalClient := tc.universalClientFn()
			regionalDetails, clientMocks := tc.regionalDetailsFn()

			defer func(t *testing.T) {
				mock.AssertExpectationsForObjects(t, directorClient, universalClient)
				for _, clientMock := range clientMocks {
					clientMock.AssertExpectations(t)
				}
			}(t)

			manager, err := NewTenantsManager(jobCfg, directorClient, universalClient, regionalDetails, tenantConverter)
			require.NoError(t, err)

			err = manager.DeleteTenants(ctx, busTenants)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenantManager_FetchTenant(t *testing.T) {
	ctx := context.TODO()

	// GIVEN
	tenantConverter := domaintenant.NewConverter()

	jobConfig := configForTenantType(tenant.Account)

	busTenant := fixBusinessTenantMappingInput("1", provider, "subdomain-1", region, "", tenant.Subaccount)
	event := fixEvent(t, "Subaccount", busTenant.Parent, eventFieldsFromTenant(tenant.Subaccount, jobConfig.UniversalClientAPIConfig.TenantFieldMapping, busTenant))

	queryParams := QueryParams{
		"pageSize": "1",
		"pageNum":  "1",
		"entityId": busTenant.ExternalTenant,
	}

	testCases := []struct {
		name              string
		jobConfigFn       func() resync.JobConfig
		region            string
		directorClientFn  func() *automock.DirectorGraphQLClient
		universalClientFn func() *automock.EventAPIClient
		regionalDetailsFn func(resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient)
		expectedTenant    *model.BusinessTenantMappingInput
		expectedErrMsg    string
	}{
		{
			name: "Success when tenant is found in the central region",
			jobConfigFn: func() resync.JobConfig {
				return configWithRegionsForSubaccounts(centralRegion, region)
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				queryParams := QueryParams{
					"pageSize": "1",
					"pageNum":  "1",
					"entityId": busTenant.ExternalTenant,
				}
				client.On("FetchTenantEventsPage", CreatedSubaccountType, queryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", UpdatedSubaccountType, queryParams).Return(nil, nil).Once()
				return client
			},
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				return nil, nil
			},
			expectedTenant: &busTenant,
		},
		{
			name: "Success when tenant is found from regional client",
			jobConfigFn: func() resync.JobConfig {
				return configWithRegionsForSubaccounts(centralRegion, region)
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", CreatedSubaccountType, queryParams).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", UpdatedSubaccountType, queryParams).Return(nil, nil).Once()
				return client
			},
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", CreatedSubaccountType, queryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", UpdatedSubaccountType, queryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[region] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[region],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
			expectedTenant: &busTenant,
		},
		{
			name: "[Temporary] Success when tenant is not found",
			jobConfigFn: func() resync.JobConfig {
				return configWithRegionsForSubaccounts(centralRegion, region)
			},
			region:           centralRegion,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			universalClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", CreatedSubaccountType, queryParams).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", UpdatedSubaccountType, queryParams).Return(nil, nil).Once()
				return client
			},
			regionalDetailsFn: func(cfg resync.JobConfig) (map[string]resync.RegionDetails, []*automock.EventAPIClient) {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", CreatedSubaccountType, queryParams).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", UpdatedSubaccountType, queryParams).Return(nil, nil).Once()

				details := make(map[string]resync.RegionDetails)
				details[region] = resync.RegionDetails{
					EventAPIRegionalConfig: *cfg.RegionalAPIConfigs[region],
					RegionalClient:         client,
				}

				return details, []*automock.EventAPIClient{client}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.jobConfigFn()

			directorClient := tc.directorClientFn()
			universalClient := tc.universalClientFn()
			regionalDetails, clientMocks := tc.regionalDetailsFn(cfg)

			defer func(t *testing.T) {
				mock.AssertExpectationsForObjects(t, directorClient, universalClient)
				for _, clientMock := range clientMocks {
					clientMock.AssertExpectations(t)
				}
			}(t)

			manager, err := NewTenantsManager(cfg, directorClient, universalClient, regionalDetails, tenantConverter)
			require.NoError(t, err)

			actualTenant, err := manager.FetchTenant(ctx, busTenant.ExternalTenant)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTenant, actualTenant)
			}
		})
	}
}

func configWithRegionsForSubaccounts(regions ...string) resync.JobConfig {
	cfg := configForTenantType(tenant.Subaccount)
	regionalCfgs := make(map[string]*EventAPIRegionalConfig)
	for _, r := range regions {
		regionalCfgs[r] = &EventAPIRegionalConfig{
			APIClientConfig:   APIClientConfig{},
			IsUniversalClient: true,
			RegionName:        r,
		}
	}
	cfg.RegionalAPIConfigs = regionalCfgs
	return cfg
}

func configForTenantType(tenantType tenant.Type) resync.JobConfig {
	return resync.JobConfig{
		EventsConfig: EventsConfig{
			QueryConfig: QueryConfig{
				PageNumField:   "pageNum",
				PageSizeField:  "pageSize",
				TimestampField: "timestamp",
				RegionField:    "region",
				PageSizeValue:  "1",
				PageStartValue: "1",
				EntityField:    "entityId",
			},
			PagingConfig: PagingConfig{
				TotalPagesField:   "pages",
				TotalResultsField: "total",
			},
			RegionalAPIConfigs: map[string]*EventAPIRegionalConfig{
				centralRegion: {
					APIClientConfig: APIClientConfig{
						APIConfig: APIConfig{},
						TenantFieldMapping: TenantFieldMapping{
							EventsField:            "events",
							NameField:              "name",
							IDField:                "id",
							GlobalAccountGUIDField: "globalAccountGUID",
							SubaccountIDField:      "subaccountId",
							CustomerIDField:        "customerId",
							SubdomainField:         "subdomain",
							DetailsField:           "eventData",
							RegionField:            "region",
							EntityIDField:          "entityId",
							EntityTypeField:        "type",
						},
						MovedSubaccountsFieldMapping: MovedSubaccountsFieldMapping{
							SubaccountID: "subaccountId",
							SourceTenant: "source",
							TargetTenant: "target",
						},
						OAuthConfig:   OAuth2Config{},
						ClientTimeout: 0,
					},
					IsUniversalClient: false,
					RegionName:        centralRegion,
				},
			},
			TenantOperationChunkSize: 500,
			RetryAttempts:            1,
		},
		ResyncConfig:   ResyncConfig{},
		KubeConfig:     KubeConfig{},
		JobName:        "tenant-fetcher",
		TenantProvider: provider,
		TenantType:     tenantType,
	}
}

func eventFieldsFromTenant(tenantType tenant.Type, tenantFieldMapping TenantFieldMapping, tenantInput model.BusinessTenantMappingInput) map[string]string {
	fields := map[string]string{
		tenantFieldMapping.IDField:        tenantInput.ExternalTenant,
		tenantFieldMapping.NameField:      tenantInput.Name,
		tenantFieldMapping.SubdomainField: tenantInput.Subdomain,
		tenantFieldMapping.RegionField:    tenantInput.Region,
	}
	switch tenantType {
	case tenant.Account:
		fields[tenantFieldMapping.EntityTypeField] = "GlobalAccount"
		fields[tenantFieldMapping.GlobalAccountGUIDField] = tenantInput.ExternalTenant
		fields[tenantFieldMapping.CustomerIDField] = tenantInput.Parent
	case tenant.Subaccount:
		fields[tenantFieldMapping.EntityTypeField] = "Subaccount"
		fields[tenantFieldMapping.SubaccountIDField] = tenantInput.ExternalTenant
		fields[tenantFieldMapping.GlobalAccountGUIDField] = tenantInput.Parent
	}
	return fields
}

func matchArrayWithoutOrderArgument(expected []graphql.BusinessTenantMappingInput) interface{} {
	return mock.MatchedBy(func(actual []graphql.BusinessTenantMappingInput) bool {
		if len(expected) != len(actual) {
			return false
		}
		matched := make([]bool, len(actual))
		for i := 0; i < len(matched); i++ {
			matched[i] = false
		}
		for i := 0; i < len(expected); i++ {
			for j := 0; j < len(actual); j++ {
				if assert.ObjectsAreEqual(expected[i], actual[j]) {
					matched[j] = true
				}
			}
		}
		for i := 0; i < len(matched); i++ {
			if matched[i] {
				continue
			}
			return false
		}
		return true
	})
}
