package destinationfetchersvc_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/tidwall/gjson"

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
	tenantID                = "f09ba084-0e82-49ab-ab2e-b7ecc988312d"
	runtimeID               = "d09ba084-0e82-49ab-ab2e-b7ecc988312d"
	destinationID           = "a447698f-acb1-4b3c-b2e7-e65bcf543538"
	appID                   = "69454c64-6284-404f-af3d-4bbbdee2c4e2"
	tenantLabelKey          = "subaccount"
	regionLabelKey          = "region"
	region                  = "region1"
	UUID                    = "9b26a428-d526-469c-a5ef-2856f3ce0430"
	subdomainLabelValue     = "127" // will be replaced in 127.0.0.1 when fetching token for destination service
	selfRegDistinguishLabel = "selfRegDistinguishLabel"
	correlationID           = "sap.s4:communicationScenario:SAP_COM_0108"
)

var (
	testErr               = errors.New("test error")
	formationAssignmentID = "c7494c4f-606a-4381-b095-8548c6726fc6"
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
		Name                    string
		LabelRepo               func() *automock.LabelRepo
		DestRepo                func() *automock.DestinationRepo
		Transactioner           func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner)
		BundleRepo              func() *automock.BundleRepo
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		UUIDService             func() *automock.UUIDService
		ExpectedErrorOutput     string
		DestServiceHandler      func(w http.ResponseWriter, r *http.Request)
	}{
		{
			Name: "Sync tenant destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              successfulBundleRepo("bundleID"),
			DestRepo:                successfulDestinationRepo("bundleID"),
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
		},
		{
			Name: "Successful sync of destinations but failing to delete old should not return error",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              successfulBundleRepo("bundleID"),
			DestRepo:                successfulInsertFailingDeleteDestinationRepo("bundleID"),
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
		},
		{
			Name: "When getting bundles fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              failingBundleRepo,
			DestRepo:                unusedDestinationsRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "When destination without system identifiers is returned should try to find bundles using the formation assignments",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              successfulBundleRepoForFormationAssignment("bundleID"),
			DestRepo:                successfulGetDeleteUpsertDestinationRepo("bundleID"),
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: successfulFormationAssignmentRepo,
			DestServiceHandler:      destinationWithoutIdentifiersHandler(t),
		},
		{
			Name: "When destination without system identifiers is returned should try to find bundles using the formation assignments but there is no formation assignment in the destination",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                successfulGetDeleteDestinationWithoutFormationAssignmentRepo("bundleID"),
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: successfulFormationAssignmentRepo,
			DestServiceHandler:      destinationWithoutIdentifiersHandler(t),
		},
		{
			Name: "When destination without system identifiers is returned and getting destination fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                failingGetDestinationRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: successfulFormationAssignmentRepo,
			DestServiceHandler:      destinationWithoutIdentifiersHandler(t),
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "When destination without system identifiers is returned and getting formation assignment fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                successfulGetDestinationRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: failingFormationAssignmentRepo,
			DestServiceHandler:      destinationWithoutIdentifiersHandler(t),
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "When destination without system identifiers is returned and listing bundles by appID and correlationID fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              failingListByApplicationAndCorrelationIDsBundleRepo,
			DestRepo:                successfulGetDestinationRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: successfulFormationAssignmentRepo,
			DestServiceHandler:      destinationWithoutIdentifiersHandler(t),
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "When destination without system identifiers is returned and upserting destination fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              successfulBundleRepoForFormationAssignment("bundle-id"),
			DestRepo:                failingUpsertDestinationRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: successfulFormationAssignmentRepo,
			DestServiceHandler:      destinationWithoutIdentifiersHandler(t),
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "When destination upsert or delete fails should stop processing destinations",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              successfulBundleRepo("bundleID"),
			DestRepo:                failingDestinationRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name:                    "Failed to begin transaction to database",
			Transactioner:           txGen.ThatFailsOnBegin,
			LabelRepo:               unusedLabelRepo,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                unusedDestinationsRepo,
			UUIDService:             unusedUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     testErr.Error(),
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
			BundleRepo:              unusedBundleRepo,
			DestRepo:                unusedDestinationsRepo,
			UUIDService:             unusedUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     fmt.Sprintf("tenant %s not found", tenantID),
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
			BundleRepo:              unusedBundleRepo,
			DestRepo:                unusedDestinationsRepo,
			UUIDService:             unusedUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name:                    "Failed to commit transaction",
			Transactioner:           txGen.ThatFailsOnCommit,
			LabelRepo:               successfulLabelSubdomainRequest,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                unusedDestinationsRepo,
			UUIDService:             unusedUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "Failed getting region",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			LabelRepo:               failedLabelRegionAndSuccessfulSubdomainRequest,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                unusedDestinationsRepo,
			UUIDService:             unusedUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
			ExpectedErrorOutput:     testErr.Error(),
		},
		{
			Name: "When destination service returns zero pages - just remove old ones",
			Transactioner: func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelRepo:               successfulLabelRegionAndSubdomainRequest,
			BundleRepo:              unusedBundleRepo,
			DestRepo:                successfulDeleteDestinationRepo,
			UUIDService:             successfulUUIDService,
			FormationAssignmentRepo: unusedFormationAssignmentRepo,
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
			formationAssignmentRepo := testCase.FormationAssignmentRepo()
			tenantRepo := unusedTenantRepo()
			defer mock.AssertExpectationsForObjects(t, tx, destRepo, labelRepo, uuidService, bundleRepo, tenantRepo)

			destSvc := destinationfetchersvc.DestinationService{
				Transactioner:           tx,
				UUIDSvc:                 uuidService,
				DestinationRepo:         destRepo,
				BundleRepo:              bundleRepo,
				LabelRepo:               labelRepo,
				TenantRepo:              tenantRepo,
				DestinationsConfig:      destConfig,
				FormationAssignmentRepo: formationAssignmentRepo,
				APIConfig:               destAPIConfig,
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
			formationAssignmentRepo := unusedFormationAssignmentRepo()
			tenantRepo := testCase.TenantRepo()

			defer mock.AssertExpectationsForObjects(t, tx, destRepo, labelRepo, uuidService, bundleRepo, tenantRepo)

			destSvc := destinationfetchersvc.NewDestinationService(tx, uuidService, destRepo, bundleRepo, labelRepo, destConfig, destAPIConfig, tenantRepo, formationAssignmentRepo, selfRegDistinguishLabel)

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

func TestService_IsTenantSubscribed(t *testing.T) {
	//GIVEN
	txGen := txtest.NewTransactionContextGenerator(testErr)
	destAPIConfig := defaultAPIConfig()
	destConfig := defaultDestinationConfig(t, "invalid")

	testCases := []struct {
		Name                string
		ExpectedTenantIDs   []string
		TenantRepo          func() *automock.TenantRepo
		Transactioner       func() (*persistenceAutomock.PersistenceTx, *persistenceAutomock.Transactioner)
		ExpectedExistence   bool
		ExpectedErrorOutput string
	}{
		{
			Name:              "Tenant is subscribed",
			ExpectedTenantIDs: []string{"a", "b"},
			TenantRepo: func() *automock.TenantRepo {
				tenantRepo := unusedTenantRepo()
				tenantRepo.On("ExistsSubscribed", mock.Anything, tenantID, selfRegDistinguishLabel).Return(true, nil).Once()
				return tenantRepo
			},
			ExpectedExistence: true,
			Transactioner:     txGen.ThatSucceeds,
		},
		{
			Name: "Tenant repo returns error",
			TenantRepo: func() *automock.TenantRepo {
				tenantRepo := unusedTenantRepo()
				tenantRepo.On("ExistsSubscribed", mock.Anything, tenantID, selfRegDistinguishLabel).Return(false, testErr).Once()
				return tenantRepo
			},
			Transactioner:       txGen.ThatSucceeds,
			ExpectedExistence:   false,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name:                "Tenant repo returns error on Begin",
			TenantRepo:          unusedTenantRepo,
			Transactioner:       txGen.ThatFailsOnBegin,
			ExpectedExistence:   false,
			ExpectedErrorOutput: testErr.Error(),
		},
		{
			Name: "Tenant repo returns error on Commit",
			TenantRepo: func() *automock.TenantRepo {
				tenantRepo := unusedTenantRepo()
				tenantRepo.On("ExistsSubscribed", mock.Anything, tenantID, selfRegDistinguishLabel).Return(true, nil).Once()
				return tenantRepo
			},
			Transactioner:       txGen.ThatFailsOnCommit,
			ExpectedExistence:   false,
			ExpectedErrorOutput: testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, tx := testCase.Transactioner()
			destRepo := unusedDestinationsRepo()
			labelRepo := unusedLabelRepo()
			uuidService := unusedUUIDService()
			bundleRepo := unusedBundleRepo()
			formationAssignmentRepo := unusedFormationAssignmentRepo()
			tenantRepo := testCase.TenantRepo()

			defer mock.AssertExpectationsForObjects(t, tx, destRepo, labelRepo, uuidService, bundleRepo, tenantRepo)

			destSvc := destinationfetchersvc.NewDestinationService(tx, uuidService, destRepo, bundleRepo, labelRepo, destConfig, destAPIConfig, tenantRepo, formationAssignmentRepo, selfRegDistinguishLabel)

			ctx := context.Background()
			// WHEN
			isTenantSubscribed, err := destSvc.IsTenantSubscribed(ctx, tenantID)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedExistence, isTenantSubscribed)
			}
		})
	}
}

func unusedLabelRepo() *automock.LabelRepo              { return &automock.LabelRepo{} }
func unusedDestinationsRepo() *automock.DestinationRepo { return &automock.DestinationRepo{} }
func unusedBundleRepo() *automock.BundleRepo            { return &automock.BundleRepo{} }
func unusedFormationAssignmentRepo() *automock.FormationAssignmentRepository {
	return &automock.FormationAssignmentRepository{}
}
func unusedTenantRepo() *automock.TenantRepo   { return &automock.TenantRepo{} }
func unusedUUIDService() *automock.UUIDService { return &automock.UUIDService{} }

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

func successfulBundleRepoForFormationAssignment(bundleID string) func() *automock.BundleRepo {
	return func() *automock.BundleRepo {
		bundleRepo := unusedBundleRepo()
		bundleRepo.On("ListByApplicationAndCorrelationIDs",
			mock.Anything, tenantID, appID, correlationID).Return(
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

func failingListByApplicationAndCorrelationIDsBundleRepo() *automock.BundleRepo {
	bundleRepo := unusedBundleRepo()
	bundleRepo.On("ListByApplicationAndCorrelationIDs",
		mock.Anything, tenantID, appID, correlationID).Return(nil, testErr)
	return bundleRepo
}

func successfulGetDeleteUpsertDestinationRepo(bundleID string) func() *automock.DestinationRepo {
	return func() *automock.DestinationRepo {
		destinationRepo := unusedDestinationsRepo()
		destinationRepo.On("GetDestinationByNameAndTenant",
			mock.Anything, destination1Name, tenantID).Return(destinationModelByJSONString(destinationID, exampleDestination1, &formationAssignmentID), nil)
		destinationRepo.On("Upsert",
			mock.Anything, mock.Anything, mock.AnythingOfType("string"), tenantID, bundleID, mock.AnythingOfType("string")).Return(nil)
		destinationRepo.On("DeleteOld",
			mock.Anything, UUID, tenantID).Return(nil)
		return destinationRepo
	}
}

func successfulGetDeleteDestinationWithoutFormationAssignmentRepo(bundleID string) func() *automock.DestinationRepo {
	return func() *automock.DestinationRepo {
		destinationRepo := unusedDestinationsRepo()
		destinationRepo.On("GetDestinationByNameAndTenant",
			mock.Anything, destination1Name, tenantID).Return(destinationModelByJSONString(destinationID, exampleDestination1, nil), nil)
		destinationRepo.On("DeleteOld",
			mock.Anything, UUID, tenantID).Return(nil)
		return destinationRepo
	}
}

func successfulGetDestinationRepo() *automock.DestinationRepo {
	destinationRepo := unusedDestinationsRepo()
	destinationRepo.On("GetDestinationByNameAndTenant",
		mock.Anything, destination1Name, tenantID).Return(destinationModelByJSONString(destinationID, exampleDestination1, &formationAssignmentID), nil)
	return destinationRepo
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

func failingGetDestinationRepo() *automock.DestinationRepo {
	destinationRepo := unusedDestinationsRepo()
	destinationRepo.On("GetDestinationByNameAndTenant",
		mock.Anything, destination1Name, tenantID).Return(nil, testErr)
	return destinationRepo
}

func failingUpsertDestinationRepo() *automock.DestinationRepo {
	destinationRepo := unusedDestinationsRepo()
	destinationRepo.On("GetDestinationByNameAndTenant",
		mock.Anything, destination1Name, tenantID).Return(destinationModelByJSONString(destinationID, exampleDestination1, &formationAssignmentID), nil)
	destinationRepo.On("Upsert",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testErr)
	return destinationRepo
}

func successfulFormationAssignmentRepo() *automock.FormationAssignmentRepository {
	formationAssignmentRepo := unusedFormationAssignmentRepo()
	formationAssignmentRepo.On("GetGlobalByID",
		mock.Anything, formationAssignmentID).Return(formationAssignmentModel(formationAssignmentID, appID), nil)
	return formationAssignmentRepo
}

func failingFormationAssignmentRepo() *automock.FormationAssignmentRepository {
	formationAssignmentRepo := unusedFormationAssignmentRepo()
	formationAssignmentRepo.On("GetGlobalByID",
		mock.Anything, formationAssignmentID).Return(nil, testErr)
	return formationAssignmentRepo
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
		tenantRepo.On("ListBySubscribedRuntimesAndApplicationTemplates", mock.Anything, selfRegDistinguishLabel).Return(tenants, nil)
		return tenantRepo
	}
}

func failingTenantRepo() *automock.TenantRepo {
	tenantRepo := unusedTenantRepo()
	tenantRepo.On("ListBySubscribedRuntimesAndApplicationTemplates", mock.Anything, selfRegDistinguishLabel).Return(nil, testErr)
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

func destinationModelByJSONString(id, destination string, formationAssignmentID *string) *model.Destination {
	return &model.Destination{
		ID:                    id,
		Name:                  gjson.Get(destination, "Name").String(),
		Type:                  gjson.Get(destination, "Type").String(),
		URL:                   gjson.Get(destination, "URL").String(),
		Authentication:        gjson.Get(destination, "Authentication").String(),
		FormationAssignmentID: formationAssignmentID,
	}
}

func formationAssignmentModel(id, targetID string) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:         id,
		Target:     targetID,
		TargetType: model.FormationAssignmentTypeApplication,
	}
}

func destinationWithoutIdentifiersHandler(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Page-Count", "0")

		_, err := w.Write([]byte(fmt.Sprintf("[%s]", exampleDestinationWithoutIdentifiers)))
		assert.NoError(t, err)
	}
}
