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
		ID: destinationSubaccountID,
	}
	correlationIDs    []string
	emptyDestinations []*model.Destination
	ctx               = context.TODO()
	testErr           = errors.New(errMsg)
)

func TestService_CreateDesignTimeDestinations(t *testing.T) {
	designTimeDestsDetails := fixDesignTimeDestinationsDetails()
	designTimeDestDetails := designTimeDestsDetails[0]

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  designTimeDestDetails.Name,
		Type:                  designTimeDestDetails.Type,
		URL:                   designTimeDestDetails.URL,
		Authentication:        designTimeDestDetails.Authentication,
		SubaccountID:          tenant.ID,
		FormationAssignmentID: &fa.ID,
	}
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.Name, tenant.ID).Return(destModel, nil)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.Name, tenant.ID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.Destination))
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when getting destination from db and error is different from 'Not Found'",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.Name, tenant.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when destination from db is found but its formation assignment id is different from the provided formation assignment",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.Name, tenant.ID).Return(destModelWithDifferentFAID, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "Could not have second destination with the same name and tenant ID but with different assignment ID",
		},
		{
			Name: "Error when creating destination via destination creator service",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth).Return(testErr)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.Name, tenant.ID).Return(destModel, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when upserting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, designTimeDestDetails.SubaccountID, &fa).Return(designTimeDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateDesignTimeDestinations", ctx, designTimeDestDetails, &fa, initialDepth).Return(nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, designTimeDestDetails.SubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("GetDestinationByNameAndTenant", ctx, designTimeDestDetails.Name, tenant.ID).Return(destModel, nil)
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
			err := svc.CreateDesignTimeDestinations(ctx, designTimeDestsDetails, &fa)

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
	basicReqBody := fixBasicReqBody()

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  basicReqBody.Name,
		Type:                  string(basicReqBody.Type),
		URL:                   basicReqBody.URL,
		Authentication:        string(basicReqBody.AuthenticationType),
		SubaccountID:          tenant.ID,
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
				destCreatorSvc.On("PrepareBasicRequestBody", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs).Return(basicReqBody, nil)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
				destCreatorSvc.On("PrepareBasicRequestBody", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs).Return(basicReqBody, nil)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth).Return(testErr)
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
			Name: "Error when preparing basic req body",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
				destCreatorSvc.On("PrepareBasicRequestBody", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs).Return(nil, testErr)
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
				destCreatorSvc.On("ValidateDestinationSubaccount", ctx, basicDestDetails.SubaccountID, &fa).Return(basicDestDetails.SubaccountID, nil)
				destCreatorSvc.On("CreateBasicCredentialDestinations", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
				destCreatorSvc.On("PrepareBasicRequestBody", ctx, basicDestDetails, basicAuthCreds, &fa, correlationIDs).Return(basicReqBody, nil)
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
			err := svc.CreateBasicCredentialDestinations(ctx, basicDestsDetails, basicAuthCreds, &fa, correlationIDs)

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

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  clientCertAuthDestDetails.Name,
		Type:                  clientCertAuthDestDetails.Type,
		URL:                   clientCertAuthTypeCreds.URL,
		Authentication:        clientCertAuthDestDetails.Authentication,
		SubaccountID:          tenant.ID,
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth).Return(testErr)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, clientCertAuthDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateClientCertificateDestination", ctx, clientCertAuthDestDetails, clientCertAuthTypeCreds, &fa, correlationIDs, initialDepth).Return(nil)
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
			err := svc.CreateClientCertificateAuthenticationDestination(ctx, clientCertAuthDestsDetails, clientCertAuthTypeCreds, &fa, correlationIDs)

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

	destModel := &model.Destination{
		ID:                    fixUUID(),
		Name:                  samlAssertionDestDetails.Name,
		Type:                  samlAssertionDestDetails.Type,
		URL:                   samlAuthCreds.URL,
		Authentication:        samlAssertionDestDetails.Authentication,
		SubaccountID:          tenant.ID,
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting tenant by external ID",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth).Return(testErr)
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
				destCreatorSvc.On("EnsureDestinationSubaccountIDsCorrectness", ctx, samlAssertionDestsDetails, &fa).Return(nil)
				destCreatorSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDestDetails, samlAuthCreds, &fa, correlationIDs, initialDepth).Return(nil)
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
			err := svc.CreateSAMLAssertionDestination(ctx, samlAssertionDestsDetails, samlAuthCreds, &fa, correlationIDs)

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
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				destCreatorSvc.On("DeleteCertificate", ctx, samlDestCertName, externalDestinationSubaccountID, destinationInstanceID, &fa).Return(nil).Once()
				destCreatorSvc.On("DeleteDestination", ctx, samlAssertionDestName, externalDestinationSubaccountID, destinationInstanceID, &fa).Return(nil).Once()
				destCreatorSvc.On("DeleteDestination", ctx, basicDestName, externalDestinationSubaccountID, destinationInstanceID, &fa).Return(nil).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByTenantIDAndAssignmentID", ctx, tenant.ID, fa.ID).Return(destinations, nil)
				destinationRepo.On("DeleteByDestinationNameAndAssignmentID", ctx, basicDestModel.Name, fa.ID, tenant.ID).Return(nil).Once()
				destinationRepo.On("DeleteByDestinationNameAndAssignmentID", ctx, samlDestModel.Name, fa.ID, tenant.ID).Return(nil).Once()
				return destinationRepo
			},
		},
		{
			Name: "Success when there are no destinations",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByTenantIDAndAssignmentID", ctx, tenant.ID, fa.ID).Return(emptyDestinations, nil)
				return destinationRepo
			},
		},
		{
			Name: "Error when getting consumer token",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return("", testErr)
				return destCreatorSvc
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when getting by external tenant",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(nil, testErr)
				return tenantRepo
			},
			ExpectedErrMessage: "while getting tenant by external ID",
		},
		{
			Name: "Error when listing destinations",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByTenantIDAndAssignmentID", ctx, tenant.ID, fa.ID).Return(nil, testErr)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when deleting certificate",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				destCreatorSvc.On("DeleteCertificate", ctx, samlDestCertName, externalDestinationSubaccountID, destinationInstanceID, &fa).Return(testErr).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByTenantIDAndAssignmentID", ctx, tenant.ID, fa.ID).Return([]*model.Destination{samlDestModel}, nil)
				return destinationRepo
			},
			ExpectedErrMessage: "while deleting destination certificate with name:",
		},
		{
			Name: "Error when deleting destination via destination creator",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				destCreatorSvc.On("DeleteDestination", ctx, basicDestModel.Name, externalDestinationSubaccountID, destinationInstanceID, &fa).Return(testErr).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByTenantIDAndAssignmentID", ctx, tenant.ID, fa.ID).Return([]*model.Destination{basicDestModel}, nil)
				return destinationRepo
			},
			ExpectedErrMessage: errMsg,
		},
		{
			Name: "Error when deleting destination in db",
			DestinationCreatorServiceFn: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("GetConsumerTenant", ctx, &fa).Return(externalDestinationSubaccountID, nil)
				destCreatorSvc.On("DeleteDestination", ctx, basicDestModel.Name, externalDestinationSubaccountID, destinationInstanceID, &fa).Return(nil).Once()
				return destCreatorSvc
			},
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", ctx, externalDestinationSubaccountID).Return(tenant, nil)
				return tenantRepo
			},
			DestinationRepoFn: func() *automock.DestinationRepository {
				destinationRepo := &automock.DestinationRepository{}
				destinationRepo.On("ListByTenantIDAndAssignmentID", ctx, tenant.ID, fa.ID).Return([]*model.Destination{basicDestModel}, nil)
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
			err := svc.DeleteDestinations(ctx, &fa)

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
