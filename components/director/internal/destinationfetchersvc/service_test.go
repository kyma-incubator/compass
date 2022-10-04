package destinationfetchersvc_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	persistenceAutomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
)

const (
	tenantID            = "f09ba084-0e82-49ab-ab2e-b7ecc988312d"
	runtimeID           = "d09ba084-0e82-49ab-ab2e-b7ecc988312d"
	tenantLabelKey      = "subaccount"
	regionLabelKey      = "region"
	region              = "region1"
	UUID                = "9b26a428-d526-469c-a5ef-2856f3ce0430"
	subdomainLabelValue = "127" // will be replaced in 127.0.0.1 when fetching token for destination service
)

var (
	testErr = errors.New("test error")
)

func TestService_SyncTenantDestinations(t *testing.T) {
	//GIVEN
	destinationServer := newDestinationServer(t)
	destinationServer.server.Start()
	defer destinationServer.server.Close()

	txGen := txtest.NewTransactionContextGenerator(testErr)
	destAPIConfig := defaultAPIConfig()
	destConfig := defaultDestinationConfig(t, destinationServer.server.URL)

	testCases := []struct {
		Name                string
		LabelRepo           func() *automock.LabelRepo
		DestRepo            func() *automock.DestinationRepo
		Transactioner       func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner)
		BundleRepo          func() *automock.BundleRepo
		UUIDService         func() *automock.UUIDService
		ExpectedErrorOutput string
		DestServiceHandler  func(w http.ResponseWriter, r *http.Request)
	}{
		{
			Name: "Sync tenant destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			LabelRepo:   successfulLabelRegionAndSubdomainRequest,
			BundleRepo:  successfulBundleRepo("bundleID"),
			DestRepo:    successfulDestinationRepo("bundleID"),
			UUIDService: successfulUUIDService,
		},
		{
			Name: "Successful sync of destinations but failing to delete old should not return error",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			LabelRepo:   successfulLabelRegionAndSubdomainRequest,
			BundleRepo:  successfulBundleRepo("bundleID"),
			DestRepo:    successfulInsertFailingDeleteDestinationRepo("bundleID"),
			UUIDService: successfulUUIDService,
		},
		{
			Name: "When getting bundles fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:           successfulLabelRegionAndSubdomainRequest,
			BundleRepo:          failingBundleRepo,
			DestRepo:            unusedDestinationsRepo,
			UUIDService:         successfulUUIDService,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "When no bundles are returned should continue to process destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			LabelRepo:   successfulLabelRegionAndSubdomainRequest,
			BundleRepo:  bundleRepoWithNoBundles,
			DestRepo:    successfulDeleteDestinationRepo,
			UUIDService: successfulUUIDService,
		},
		{
			Name: "When destination upsert or delete fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:           successfulLabelRegionAndSubdomainRequest,
			BundleRepo:          successfulBundleRepo("bundleID"),
			DestRepo:            failingDestinationRepo,
			UUIDService:         successfulUUIDService,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:                "Failed to begin transaction to database",
			Transactioner:       txGen.ThatFailsOnBegin,
			LabelRepo:           unusedLabelRepo,
			BundleRepo:          unusedBundleRepo,
			DestRepo:            unusedDestinationsRepo,
			UUIDService:         unusedUUIDService,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:          "Failed to find subdomain label",
			Transactioner: txGen.ThatSucceeds,
			LabelRepo: func() *automock.LabelRepo {
				repo := &automock.LabelRepo{}
				repo.On("GetSubdomainLabelForSubscribedRuntime", mock.Anything, tenantID).
					Return(nil, apperrors.NewNotFoundError(resource.Label, "id"))
				return repo
			},
			BundleRepo:          unusedBundleRepo,
			DestRepo:            unusedDestinationsRepo,
			UUIDService:         unusedUUIDService,
			ExpectedErrorOutput: fmt.Sprintf("tenant %s not found", tenantID),
		},
		{
			Name:          "Error while getting subdomain label",
			Transactioner: txGen.ThatSucceeds,
			LabelRepo: func() *automock.LabelRepo {
				repo := &automock.LabelRepo{}
				repo.On("GetSubdomainLabelForSubscribedRuntime", mock.Anything, tenantID).
					Return(nil, testErr)
				return repo
			},
			BundleRepo:          unusedBundleRepo,
			DestRepo:            unusedDestinationsRepo,
			UUIDService:         unusedUUIDService,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:                "Failed to commit transaction",
			Transactioner:       txGen.ThatFailsOnCommit,
			LabelRepo:           successfulLabelSubdomainRequest,
			BundleRepo:          unusedBundleRepo,
			DestRepo:            unusedDestinationsRepo,
			UUIDService:         unusedUUIDService,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "Failed getting region",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			LabelRepo:           failedLabelRegionAndSuccessfulSubdomainRequest,
			BundleRepo:          unusedBundleRepo,
			DestRepo:            unusedDestinationsRepo,
			UUIDService:         unusedUUIDService,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "When destination service returns only invalid destinations - do not store them and remove old ones",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			LabelRepo:   successfulLabelRegionAndSubdomainRequest,
			BundleRepo:  unusedBundleRepo,
			DestRepo:    successfulDeleteDestinationRepo,
			UUIDService: successfulUUIDService,
			DestServiceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Page-Count", "1")

				var (
					invalidDest1 = `{
						"Name": "badS4URL",
						"URL": "invalid",
						"XFSystemName": "Test S4HANA system",
						"product.name": "SAP S/4HANA Cloud",
						"communicationScenarioId": "SAP_COM_0108"
					}`
					invalidDest2 = `{
						"Name": "no URL",
						"XFSystemName": "Test S4HANA system",
						"product.name": "SAP S/4HANA Cloud",
						"x-correlation-id": "correlation-id"
					}`
					invalidDest3 = `{
						"Name": "bad URL",
						"URL": ":invalidURL",
						"x-system-name": "Test S4HANA system",
						"x-system-type": "SAP S/4HANA Cloud",
						"x-correlation-id": "correlation-id"
					}`
					invalidDest4 = `{
						"Name": "no correlation id",
						"x-system-type": "systemType",
						"x-system-id": "systemId"
					}`
				)
				_, err := w.Write([]byte(fmt.Sprintf("[%s, %s, %s, %s]",
					invalidDest1, invalidDest2, invalidDest3, invalidDest4)))
				assert.NoError(t, err)
			},
		},
		{
			Name: "When destination service returns zero pages - just remove old ones",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:   successfulLabelRegionAndSubdomainRequest,
			BundleRepo:  unusedBundleRepo,
			DestRepo:    successfulDeleteDestinationRepo,
			UUIDService: successfulUUIDService,
			DestServiceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Page-Count", "0")

				_, err := w.Write([]byte("[]"))
				assert.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			destinationServer.handler.customTenantDestinationHandler = testCase.DestServiceHandler
			defer func() {
				destinationServer.handler.customTenantDestinationHandler = nil
			}()
			_, tx := testCase.Transactioner()
			destRepo := testCase.DestRepo()
			labelRepo := testCase.LabelRepo()
			bundleRepo := testCase.BundleRepo()
			uuidService := testCase.UUIDService()
			tenantRepo := unusedTenantRepo()
			defer mock.AssertExpectationsForObjects(t, tx, destRepo, labelRepo, uuidService, bundleRepo, tenantRepo)

			destSvc := destinationfetchersvc.DestinationService{
				Transactioner:      tx,
				UUIDSvc:            uuidService,
				DestinationRepo:    destRepo,
				BundleRepo:         bundleRepo,
				LabelRepo:          labelRepo,
				TenantRepo:         tenantRepo,
				DestinationsConfig: destConfig,
				APIConfig:          destAPIConfig,
			}

			ctx := context.Background()
			// WHEN
			err := destSvc.SyncTenantDestinations(ctx, tenantID)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_FetchDestinationsSensitiveData(t *testing.T) {
	//GIVEN
	destinationServer := newDestinationServer(t)
	destinationServer.server.Start()
	defer destinationServer.server.Close()

	txGen := txtest.NewTransactionContextGenerator(testErr)
	destAPIConfig := defaultAPIConfig()
	destConfig := defaultDestinationConfig(t, destinationServer.server.URL)

	testCases := []struct {
		Name                string
		DestinationNames    []string
		TenantID            string
		LabelRepo           func() *automock.LabelRepo
		Transactioner       func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner)
		ExpectedErrorOutput string
	}{
		{
			Name:             "Fetch with empty destination list",
			DestinationNames: []string{},
			TenantID:         tenantID,
			LabelRepo:        successfulLabelRegionAndSubdomainRequest,
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
		},
		{
			Name:             "Fetch with existing destinations list",
			DestinationNames: []string{"dest1", "dest2"},
			TenantID:         tenantID,
			LabelRepo:        successfulLabelRegionAndSubdomainRequest,
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
		},
		{
			Name:             "Fetch with one non-existing destination",
			DestinationNames: []string{"dest1", "missing"},
			TenantID:         tenantID,
			LabelRepo:        successfulLabelRegionAndSubdomainRequest,
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			ExpectedErrorOutput: "Object not found",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, tx := testCase.Transactioner()
			destRepo := unusedDestinationsRepo()
			labelRepo := testCase.LabelRepo()
			uuidService := unusedUUIDService()
			bundleRepo := unusedBundleRepo()
			tenantRepo := unusedTenantRepo()

			defer mock.AssertExpectationsForObjects(t, tx, destRepo, labelRepo, uuidService, bundleRepo, tenantRepo)

			destSvc := destinationfetchersvc.DestinationService{
				Transactioner:      tx,
				UUIDSvc:            uuidService,
				DestinationRepo:    destRepo,
				BundleRepo:         bundleRepo,
				TenantRepo:         tenantRepo,
				LabelRepo:          labelRepo,
				DestinationsConfig: destConfig,
				APIConfig:          destAPIConfig,
			}

			ctx := context.Background()
			// WHEN
			resp, err := destSvc.FetchDestinationsSensitiveData(ctx, testCase.TenantID, testCase.DestinationNames)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				require.NoError(t, err)
				var parsedResponse map[string]map[string]interface{}
				require.NoError(t, json.Unmarshal(resp, &parsedResponse))
				destinations := parsedResponse["destinations"]
				require.NotNil(t, destinations)
				for _, expectedDestinationName := range testCase.DestinationNames {
					require.Contains(t, destinations, expectedDestinationName)
				}
			}
		})
	}
}

func TestService_GetSubscribedTenantIDs(t *testing.T) {
	//GIVEN
	txGen := txtest.NewTransactionContextGenerator(testErr)
	destAPIConfig := defaultAPIConfig()
	destConfig := defaultDestinationConfig(t, "invalid")

	testCases := []struct {
		Name                string
		ExpectedTenantIDs   []string
		TenantRepo          func() *automock.TenantRepo
		Transactioner       func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner)
		ExpectedErrorOutput string
	}{
		{
			Name:              "Fetch subscribed tenants",
			ExpectedTenantIDs: []string{"a", "b"},
			TenantRepo:        successfulTenantRepo([]string{"b", "a"}),
			Transactioner:     txGen.ThatSucceeds,
		},
		{
			Name:                "Tenant repo returns error",
			TenantRepo:          failingTenantRepo,
			Transactioner:       txGen.ThatSucceeds,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:                "Tenant repo returns error on Begin",
			TenantRepo:          unusedTenantRepo,
			Transactioner:       txGen.ThatFailsOnBegin,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:                "Tenant repo returns error on Commit",
			TenantRepo:          successfulTenantRepo([]string{}),
			Transactioner:       txGen.ThatFailsOnCommit,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:              "Tenant repo returns no tenants",
			ExpectedTenantIDs: []string{},
			TenantRepo:        successfulTenantRepo([]string{}),
			Transactioner:     txGen.ThatSucceeds,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, tx := testCase.Transactioner()
			destRepo := unusedDestinationsRepo()
			labelRepo := unusedLabelRepo()
			uuidService := unusedUUIDService()
			bundleRepo := unusedBundleRepo()
			tenantRepo := testCase.TenantRepo()

			defer mock.AssertExpectationsForObjects(t, tx, destRepo, labelRepo, uuidService, bundleRepo, tenantRepo)

			destSvc := destinationfetchersvc.DestinationService{
				Transactioner:      tx,
				UUIDSvc:            uuidService,
				DestinationRepo:    destRepo,
				BundleRepo:         bundleRepo,
				TenantRepo:         tenantRepo,
				LabelRepo:          labelRepo,
				DestinationsConfig: destConfig,
				APIConfig:          destAPIConfig,
			}

			ctx := context.Background()
			// WHEN
			tenantIDs, err := destSvc.GetSubscribedTenantIDs(ctx)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				require.NoError(t, err)
				for _, expectedTenantID := range testCase.ExpectedTenantIDs {
					require.Contains(t, tenantIDs, expectedTenantID)
				}
			}
		})
	}
}

func unusedLabelRepo() *automock.LabelRepo              { return &automock.LabelRepo{} }
func unusedDestinationsRepo() *automock.DestinationRepo { return &automock.DestinationRepo{} }
func unusedBundleRepo() *automock.BundleRepo            { return &automock.BundleRepo{} }
func unusedTenantRepo() *automock.TenantRepo            { return &automock.TenantRepo{} }
func unusedUUIDService() *automock.UUIDService          { return &automock.UUIDService{} }

func successfulUUIDService() *automock.UUIDService {
	uuidService := &automock.UUIDService{}
	uuidService.On("Generate").Return(UUID)
	return uuidService
}

func successfulLabelSubdomainRequest() *automock.LabelRepo {
	repo := &automock.LabelRepo{}
	label := model.NewLabelForRuntime(runtimeID, tenantID, tenantLabelKey, subdomainLabelValue)
	tid := tenantID
	label.Tenant = &tid
	repo.On("GetSubdomainLabelForSubscribedRuntime", mock.Anything, tenantID).Return(label, nil)
	return repo
}

func failedLabelRegionAndSuccessfulSubdomainRequest() *automock.LabelRepo {
	repo := &automock.LabelRepo{}
	label := model.NewLabelForRuntime(runtimeID, tenantID, tenantLabelKey, subdomainLabelValue)
	tid := tenantID
	label.Tenant = &tid
	repo.On("GetSubdomainLabelForSubscribedRuntime", mock.Anything, tenantID).Return(label, nil)
	repo.On("GetByKey", mock.Anything, tenantID, model.TenantLabelableObject, tenantID, regionLabelKey).
		Return(nil, testErr)
	return repo
}

func successfulLabelRegionAndSubdomainRequest() *automock.LabelRepo {
	repo := &automock.LabelRepo{}
	label := model.NewLabelForRuntime(runtimeID, tenantID, tenantLabelKey, subdomainLabelValue)
	tid := tenantID
	label.Tenant = &tid
	repo.On("GetSubdomainLabelForSubscribedRuntime", mock.Anything, tenantID).Return(label, nil)
	label = model.NewLabelForRuntime(runtimeID, tenantID, regionLabelKey, region)
	label.Tenant = &tid
	repo.On("GetByKey", mock.Anything, tenantID, model.TenantLabelableObject, tenantID, regionLabelKey).
		Return(label, nil)
	return repo
}

func successfulBundleRepo(bundleID string) func() *automock.BundleRepo {
	return func() *automock.BundleRepo {
		bundleRepo := unusedBundleRepo()
		bundleRepo.On("ListByDestination",
			mock.Anything, mock.Anything, mock.Anything).Return(
			[]*model.Bundle{{
				BaseEntity: &model.BaseEntity{
					ID: bundleID,
				},
			}}, nil)
		return bundleRepo
	}
}

func failingBundleRepo() *automock.BundleRepo {
	bundleRepo := unusedBundleRepo()
	bundleRepo.On("ListByDestination",
		mock.Anything, mock.Anything, mock.Anything).Return(nil, testErr)
	return bundleRepo
}

func bundleRepoWithNoBundles() *automock.BundleRepo {
	bundleRepo := unusedBundleRepo()
	bundleRepo.On("ListByDestination",
		mock.Anything, mock.Anything, mock.Anything).Return([]*model.Bundle{}, nil)
	return bundleRepo
}

func successfulDeleteDestinationRepo() *automock.DestinationRepo {
	destinationRepo := unusedDestinationsRepo()
	destinationRepo.On("DeleteOld",
		mock.Anything, UUID, tenantID).Return(nil)
	return destinationRepo
}

func successfulDestinationRepo(bundleID string) func() *automock.DestinationRepo {
	return func() *automock.DestinationRepo {
		destinationRepo := unusedDestinationsRepo()
		destinationRepo.On("Upsert",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, bundleID, mock.Anything).Return(nil)
		destinationRepo.On("DeleteOld",
			mock.Anything, UUID, tenantID).Return(nil)
		return destinationRepo
	}
}

func successfulInsertFailingDeleteDestinationRepo(bundleID string) func() *automock.DestinationRepo {
	return func() *automock.DestinationRepo {
		destinationRepo := unusedDestinationsRepo()
		destinationRepo.On("Upsert",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, bundleID, mock.Anything).Return(nil)
		destinationRepo.On("DeleteOld",
			mock.Anything, UUID, tenantID).Return(testErr)
		return destinationRepo
	}
}

func failingDestinationRepo() *automock.DestinationRepo {
	destinationRepo := unusedDestinationsRepo()
	destinationRepo.On("Upsert",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testErr)
	return destinationRepo
}

func successfulTenantRepo(tenantIDs []string) func() *automock.TenantRepo {
	return func() *automock.TenantRepo {
		tenantRepo := unusedTenantRepo()
		tenants := make([]*model.BusinessTenantMapping, 0, len(tenantIDs))
		for _, tenantID := range tenantIDs {
			tenants = append(tenants, &model.BusinessTenantMapping{
				ID: tenantID,
			})
		}
		tenantRepo.On("ListBySubscribedRuntimes", mock.Anything).Return(tenants, nil)
		return tenantRepo
	}
}

func failingTenantRepo() *automock.TenantRepo {
	tenantRepo := unusedTenantRepo()
	tenantRepo.On("ListBySubscribedRuntimes", mock.Anything).Return(nil, testErr)
	return tenantRepo
}

func defaultAPIConfig() destinationfetchersvc.DestinationServiceAPIConfig {
	return destinationfetchersvc.DestinationServiceAPIConfig{
		GoroutineLimit:                2,
		RetryInterval:                 0,
		RetryAttempts:                 2,
		EndpointGetTenantDestinations: "/subaccountDestinations",
		EndpointFindDestination:       "/destinations",
		Timeout:                       time.Second * 10,
		PageSize:                      1,
		PagingPageParam:               "$page",
		PagingSizeParam:               "$pageSize",
		PagingCountParam:              "$pageCount",
		PagingCountHeader:             "Page-Count",
		OAuthTokenPath:                "/oauth/token",
	}
}

func defaultDestinationConfig(t *testing.T, destinationServerURL string) config.DestinationsConfig {
	cert, key := generateTestCertAndKey(t, "test")
	instanceConfig := config.InstanceConfig{
		ClientID:     tenantID,
		ClientSecret: "secret",
		URL:          destinationServerURL,
		TokenURL:     destinationServerURL,
		Cert:         string(cert),
		Key:          string(key),
	}
	return config.DestinationsConfig{
		RegionToInstanceConfig: map[string]config.InstanceConfig{
			region: instanceConfig,
		},
	}
}
