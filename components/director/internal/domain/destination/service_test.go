package destination_test

import (
	"context"
	"fmt"
	"testing"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const errMsg = "test err"

var (
	fa = model.FormationAssignment{
		ID: destinationFormationAssignmentID,
	}
	tenant = &model.BusinessTenantMapping{
		ID:             internalDestinationSubaccountID,
		ExternalTenant: externalDestinationSubaccountID,
	}
	correlationIDs    []string
	emptyDestinations []*model.Destination
	ctx               = context.TODO()
	testErr           = errors.New(errMsg)
)

func TestService_CreateDesignTimeDestinations(t *testing.T) {
	designTimeDestsDetails := fixDesignTimeDestinationsDetails()
	designTimeDestDetails := designTimeDestsDetails[0]

	destModel, err := designTimeDestDetails.ToModelDestination(fixUUID(), tenant.ID, fa.ID)
	require.NoError(t, err)

	destModelWithDifferentFAID := fixDestinationModelWithAuthnAndFAID(destinationName, string(destinationcreatorpkg.AuthTypeNoAuth), secondDestinationFormationAssignmentID)

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.GetName(), tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Success when there is no destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.GetName(), tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Error when validating destination subaccount",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.GetName(), tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.GetName(), tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth, false).Return(testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.GetName(), tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when upserting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, designTimeDestDetails.GetSubaccountID(), &fa, false).Return(designTimeDestDetails.GetSubaccountID(), nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.GetSubaccountID()).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.GetName(), tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(testErr)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
			ExpectedErrMessage: "while upserting design time destination with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}

			uidSvc := unusedUIDService()
			if testCase.UIDServiceFn != nil {
				uidSvc = testCase.UIDServiceFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo, uidSvc)

			svc := destination.NewService(nil, destRepo, tntRepo, uidSvc, destCreatorSvc)

			// WHEN
			err := svc.CreateDesignTimeDestinations(ctx, designTimeDestsDetails, &fa, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_CreateBasicCredentialDestinations(t *testing.T) {
	basicDestsDetails := fixBasicDestinationsDetails()
	basicDestDetails := basicDestsDetails[0]
	basicDestInfo := fixBasicDestInfo()

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  basicDestDetails.Name,
		Type:                  string(basicDestInfo.Type),
		URL:                   basicDestInfo.URL,
		Authentication:        string(basicDestInfo.AuthenticationType),
		SubaccountID:          tenant.ID,
		InstanceID:            &basicDestDetails.InstanceID,
		FormationAssignmentID: &fa.ID,
	}
	destModelWithDifferentFAID := fixDestinationModelWithAuthnAndFAID(basicDestName, string(destinationcreatorpkg.AuthTypeBasic), secondDestinationFormationAssignmentID)
	basicAuthCreds := fixBasicAuthn()

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth, false).Return(basicDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, basicDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Success when there is no destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth, false).Return(basicDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, basicDestDetails.Name, tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Error when validating destination subaccount",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, basicDestDetails.Name, tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, basicDestDetails.Name, tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth, false).Return(nil, testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, basicDestDetails.Name, tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when upserting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa, false).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth, false).Return(basicDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, basicDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, basicDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(testErr)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
			ExpectedErrMessage: "while upserting basic destination with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}

			uidSvc := unusedUIDService()
			if testCase.UIDServiceFn != nil {
				uidSvc = testCase.UIDServiceFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo, uidSvc)

			svc := destination.NewService(nil, destRepo, tntRepo, uidSvc, destCreatorSvc)

			// WHEN
			err := svc.CreateBasicCredentialDestinations(ctx, basicDestsDetails, basicAuthCreds, &fa, correlationIDs, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_CreateClientCertificateAuthenticationDestination(t *testing.T) {
	clientCertAuthDestsDetails := fixClientCertAuthDestinationsDetails()
	clientCertAuthDestDetails := clientCertAuthDestsDetails[0]
	clientCertAuthTypeCreds := fixClientCertAuthTypeAuthentication()
	clientCertDestInfo := fixClientCertDestInfo()

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  clientCertAuthDestDetails.Name,
		Type:                  string(clientCertDestInfo.Type),
		URL:                   clientCertDestInfo.URL,
		Authentication:        string(clientCertDestInfo.AuthenticationType),
		SubaccountID:          tenant.ID,
		InstanceID:            &clientCertAuthDestDetails.InstanceID,
		FormationAssignmentID: &fa.ID,
	}
	destModelWithDifferentFAID := fixDestinationModelWithAuthnAndFAID(clientCertAuthDestName, string(destinationcreatorpkg.AuthTypeClientCertificate), secondDestinationFormationAssignmentID)

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth, false).Return(clientCertDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, clientCertAuthDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Success when there is no destination in DB",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth, false).Return(clientCertDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, clientCertAuthDestDetails.Name, tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Error when ensuring destination subaccount correctness fails",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, clientCertAuthDestDetails.Name, tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, clientCertAuthDestDetails.Name, tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth, false).Return(nil, testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, clientCertAuthDestDetails.Name, tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when upserting destination in DB",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth, false).Return(clientCertDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, clientCertAuthDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, clientCertAuthDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(testErr)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
			ExpectedErrMessage: "while upserting SAML Assertion destination with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}

			uidSvc := unusedUIDService()
			if testCase.UIDServiceFn != nil {
				uidSvc = testCase.UIDServiceFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo, uidSvc)

			svc := destination.NewService(nil, destRepo, tntRepo, uidSvc, destCreatorSvc)

			// WHEN
			err := svc.CreateClientCertificateAuthenticationDestination(ctx, clientCertAuthDestsDetails, clientCertAuthTypeCreds, &fa, correlationIDs, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_CreateSAMLAssertionDestinations(t *testing.T) {
	samlAssertionDestsDetails := fixSAMLAssertionDestinationsDetails()
	samlAssertionDestDetails := samlAssertionDestsDetails[0]
	samlAuthCreds := fixSAMLAssertionAuthentication()
	samlDestInfo := fixSAMLDestInfo()

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  samlAssertionDestDetails.Name,
		Type:                  string(samlDestInfo.Type),
		URL:                   samlDestInfo.URL,
		Authentication:        string(samlDestInfo.AuthenticationType),
		SubaccountID:          tenant.ID,
		InstanceID:            &samlAssertionDestDetails.InstanceID,
		FormationAssignmentID: &fa.ID,
	}
	destModelWithDifferentFAID := fixDestinationModelWithAuthnAndFAID(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), secondDestinationFormationAssignmentID)

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth, false).Return(samlDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, samlAssertionDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Success when there is no destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth, false).Return(samlDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, samlAssertionDestDetails.Name, tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Error when validating destination subaccount",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, samlAssertionDestDetails.Name, tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, samlAssertionDestDetails.Name, tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth, false).Return(nil, testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, samlAssertionDestDetails.Name, tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when upserting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa, false).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth, false).Return(samlDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, samlAssertionDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, samlAssertionDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(testErr)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
			ExpectedErrMessage: "while upserting SAML Assertion destination with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}

			uidSvc := unusedUIDService()
			if testCase.UIDServiceFn != nil {
				uidSvc = testCase.UIDServiceFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo, uidSvc)

			svc := destination.NewService(nil, destRepo, tntRepo, uidSvc, destCreatorSvc)

			// WHEN
			err := svc.CreateSAMLAssertionDestination(ctx, samlAssertionDestsDetails, samlAuthCreds, &fa, correlationIDs, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_CreateOAuth2ClientCredentialsDestinations(t *testing.T) {
	oauth2ClientCredsDestsDetails := fixOAuth2ClientCredsDestinationsDetails()
	oauth2ClientCredsDestDetails := oauth2ClientCredsDestsDetails[0]
	oauth2ClientCredsDestInfo := fixOAuth2ClientCredsDestInfo()

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  oauth2ClientCredsDestDetails.Name,
		Type:                  string(oauth2ClientCredsDestInfo.Type),
		URL:                   oauth2ClientCredsDestInfo.URL,
		Authentication:        string(oauth2ClientCredsDestInfo.AuthenticationType),
		SubaccountID:          tenant.ID,
		InstanceID:            &oauth2ClientCredsDestDetails.InstanceID,
		FormationAssignmentID: &fa.ID,
	}
	destModelWithDifferentFAID := fixDestinationModelWithAuthnAndFAID(oauth2ClientCredsDestName, string(destinationcreatorpkg.AuthTypeOAuth2ClientCredentials), secondDestinationFormationAssignmentID)
	oaut2ClientCreds := fixOAuth2ClientCredsAuthn()

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2ClientCredentialsDestinations", ctx, oauth2ClientCredsDestDetails, oaut2ClientCreds, &fa, correlationIDs, initialDepth, false).Return(oauth2ClientCredsDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2ClientCredsDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Success when there is no destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2ClientCredentialsDestinations", ctx, oauth2ClientCredsDestDetails, oaut2ClientCreds, &fa, correlationIDs, initialDepth, false).Return(oauth2ClientCredsDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2ClientCredsDestDetails.Name, tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Error when validating destination subaccount",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2ClientCredsDestDetails.Name, tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2ClientCredsDestDetails.Name, tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2ClientCredentialsDestinations", ctx, oauth2ClientCredsDestDetails, oaut2ClientCreds, &fa, correlationIDs, initialDepth, false).Return(nil, testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2ClientCredsDestDetails.Name, tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when upserting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2ClientCredsDestDetails.SubaccountID, &fa, false).Return(oauth2ClientCredsDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2ClientCredentialsDestinations", ctx, oauth2ClientCredsDestDetails, oaut2ClientCreds, &fa, correlationIDs, initialDepth, false).Return(oauth2ClientCredsDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2ClientCredsDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2ClientCredsDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(testErr)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
			ExpectedErrMessage: "while upserting oauth2 client creds destination with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}

			uidSvc := unusedUIDService()
			if testCase.UIDServiceFn != nil {
				uidSvc = testCase.UIDServiceFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo, uidSvc)

			svc := destination.NewService(nil, destRepo, tntRepo, uidSvc, destCreatorSvc)

			// WHEN
			err := svc.CreateOAuth2ClientCredentialsDestinations(ctx, oauth2ClientCredsDestsDetails, oaut2ClientCreds, &fa, correlationIDs, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_CreateOAuth2mTLSDestinations(t *testing.T) {
	oauth2mTLSDestsDetails := fixOAuth2mTLSDestinationsDetails()
	oauth2mTLSDestDetails := oauth2mTLSDestsDetails[0]
	oauth2mTLSDestInfo := fixOAuth2mTLSDestInfo()

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  oauth2mTLSDestDetails.Name,
		Type:                  string(oauth2mTLSDestInfo.Type),
		URL:                   oauth2mTLSDestInfo.URL,
		Authentication:        string(oauth2mTLSDestInfo.AuthenticationType),
		SubaccountID:          tenant.ID,
		InstanceID:            &oauth2mTLSDestDetails.InstanceID,
		FormationAssignmentID: &fa.ID,
	}
	destModelWithDifferentFAID := fixDestinationModelWithAuthnAndFAID(oauth2mTLSDestName, string(destinationcreatorpkg.AuthTypeOAuth2mTLS), secondDestinationFormationAssignmentID)
	oauth2mTLSAuthn := fixOAuth2mTLSAuthn()

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2mTLSDestinations", ctx, oauth2mTLSDestDetails, oauth2mTLSAuthn, &fa, correlationIDs, initialDepth, false).Return(oauth2mTLSDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2mTLSDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Success when there is no destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2mTLSDestinations", ctx, oauth2mTLSDestDetails, oauth2mTLSAuthn, &fa, correlationIDs, initialDepth, false).Return(oauth2mTLSDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2mTLSDestDetails.Name, tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(nil)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
		},
		{
			Name: "Error when validating destination subaccount",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2mTLSDestDetails.Name, tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2mTLSDestDetails.Name, tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2mTLSDestinations", ctx, oauth2mTLSDestDetails, oauth2mTLSAuthn, &fa, correlationIDs, initialDepth, false).Return(nil, testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2mTLSDestDetails.Name, tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when upserting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DetermineDestinationSubaccount", ctx, oauth2mTLSDestDetails.SubaccountID, &fa, false).Return(oauth2mTLSDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateOAuth2mTLSDestinations", ctx, oauth2mTLSDestDetails, oauth2mTLSAuthn, &fa, correlationIDs, initialDepth, false).Return(oauth2mTLSDestInfo, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, oauth2mTLSDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, oauth2mTLSDestDetails.Name, tenant.ID).Return(destModel, nil)
				destinationRepo.On("UpsertWithEmbeddedTenant", ctx, destModel).Return(testErr)
				return destinationRepo
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(fixUUID())
				return uidSvc
			},
			ExpectedErrMessage: "while upserting oauth2 mTLS destination with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}

			uidSvc := unusedUIDService()
			if testCase.UIDServiceFn != nil {
				uidSvc = testCase.UIDServiceFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo, uidSvc)

			svc := destination.NewService(nil, destRepo, tntRepo, uidSvc, destCreatorSvc)

			// WHEN
			err := svc.CreateOAuth2mTLSDestinations(ctx, oauth2mTLSDestsDetails, oauth2mTLSAuthn, &fa, correlationIDs, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestService_DeleteDestinations(t *testing.T) {
	basicDestModel := fixDestinationModelWithAuthnAndFAID(basicDestName, string(destinationcreatorpkg.AuthTypeBasic), fa.ID)
	samlDestModel := fixDestinationModelWithAuthnAndFAID(samlAssertionDestName, string(destinationcreatorpkg.AuthTypeSAMLAssertion), fa.ID)
	destinations := []*model.Destination{basicDestModel, samlDestModel}

	samlDestCertName := fmt.Sprintf("%s-%s", destinationcreatorpkg.AuthTypeSAMLAssertion, destinationFormationAssignmentID)

	testCases := []struct {
		Name                        string
		DestinationCreatorServiceFn func() *automock.DestinationCreatorService
		TenantRepoFn                func() *automock.TenantRepository
		DestinationRepoFn           func() *automock.DestinationRepository
		ExpectedErrMessage          string
	}{
		{
			Name: "Success",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DeleteCertificate", ctx, samlDestCertName, externalDestinationSubaccountID, destinationInstanceID, &fa, false).Return(nil).Once()
				destCreatorSvc.On("DeleteDestination", ctx, samlAssertionDestName, externalDestinationSubaccountID, destinationInstanceID, &fa, false).Return(nil).Once()
				destCreatorSvc.On("DeleteDestination", ctx, basicDestName, externalDestinationSubaccountID, destinationInstanceID, &fa, false).Return(nil).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, internalDestinationSubaccountID).Return(tenant, nil).Twice()
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return(destinations, nil)
				destinationRepo.On("DeleteByDestinationNameAndAssignmentID", ctx, basicDestModel.Name, fa.ID, tenant.ID).Return(nil).Once()
				destinationRepo.On("DeleteByDestinationNameAndAssignmentID", ctx, samlDestModel.Name, fa.ID, tenant.ID).Return(nil).Once()
				return destinationRepo
			},
		},
		{
			Name: "Success when there are no destinations",
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return(emptyDestinations, nil)
				return destinationRepo
			},
		},
		{
			Name: "Error when listing destinations",
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: fmt.Sprintf("while listing destinations by assignment ID: %q: %s", destinationFormationAssignmentID, testErr.Error()),
		},
		{
			Name: "Error when getting by tenant",
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return(destinations, nil)
				return destinationRepo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, internalDestinationSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			ExpectedErrMessage: fmt.Sprintf("while getting tenant for destination subaccount ID: %q: %s", internalDestinationSubaccountID, testErr.Error()),
		},
		{
			Name: "Error when deleting certificate",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DeleteCertificate", ctx, samlDestCertName, externalDestinationSubaccountID, destinationInstanceID, &fa, false).Return(testErr).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, internalDestinationSubaccountID).Return(tenant, nil).Once()
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return([]*model.Destination{samlDestModel}, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "while deleting destination certificate with name:",
		},
		{
			Name: "Error when deleting destination via destination creator",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DeleteDestination", ctx, basicDestModel.Name, externalDestinationSubaccountID, destinationInstanceID, &fa, false).Return(testErr).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, internalDestinationSubaccountID).Return(tenant, nil).Once()
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return([]*model.Destination{basicDestModel}, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when deleting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("DeleteDestination", ctx, basicDestModel.Name, externalDestinationSubaccountID, destinationInstanceID, &fa, false).Return(nil).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", ctx, internalDestinationSubaccountID).Return(tenant, nil).Once()
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByAssignmentID", ctx, fa.ID).Return([]*model.Destination{basicDestModel}, nil)
				destinationRepo.On("DeleteByDestinationNameAndAssignmentID", ctx, basicDestModel.Name, fa.ID, tenant.ID).Return(testErr).Once()
				return destinationRepo
			},
			ExpectedErrMessage: "while deleting destination(s) by name:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorServiceFn != nil {
				destCreatorSvc = testCase.DestinationCreatorServiceFn()
			}

			tntRepo := unusedTenantRepository()
			if testCase.TenantRepoFn != nil {
				tntRepo = testCase.TenantRepoFn()
			}

			destRepo := unusedDestinationRepository()
			if testCase.DestinationRepoFn != nil {
				destRepo = testCase.DestinationRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, destCreatorSvc, tntRepo, destRepo)

			svc := destination.NewService(nil, destRepo, tntRepo, nil, destCreatorSvc)

			// WHEN
			err := svc.DeleteDestinations(ctx, &fa, false)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}
