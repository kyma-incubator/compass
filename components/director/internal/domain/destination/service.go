package destination

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator"
	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=destinationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type destinationRepository interface {
	GetDestinationByNameAndTenant(ctx context.Context, destinationName, tenantID string) (*model.Destination, error)
	DeleteByDestinationNameAndAssignmentID(ctx context.Context, destinationName, formationAssignmentID, tenantID string) error
	ListByTenantIDAndAssignmentID(ctx context.Context, tenantID, formationAssignmentID string) ([]*model.Destination, error)
	UpsertWithEmbeddedTenant(ctx context.Context, destination *model.Destination) error
}

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

//go:generate mockery --exported --name=destinationCreatorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type destinationCreatorService interface {
	CreateDesignTimeDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *model.FormationAssignment, depth uint8) error
	CreateBasicCredentialDestinations(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error
	CreateSAMLAssertionDestination(ctx context.Context, destinationDetails operators.Destination, samlAssertionAuthCreds *operators.SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error
	CreateClientCertificateDestination(ctx context.Context, destinationDetails operators.Destination, clientCertAuthCreds *operators.ClientCertAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error
	DeleteDestination(ctx context.Context, destinationName, externalDestSubaccountID, instanceID string, formationAssignment *model.FormationAssignment) error
	DeleteCertificate(ctx context.Context, certificateName, externalDestSubaccountID, instanceID string, formationAssignment *model.FormationAssignment) error
	ValidateDestinationSubaccount(ctx context.Context, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) (string, error)
	PrepareBasicRequestBody(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) (*destinationcreator.BasicAuthDestinationRequestBody, error)
	GetConsumerTenant(ctx context.Context, formationAssignment *model.FormationAssignment) (string, error)
	EnsureDestinationSubaccountIDsCorrectness(ctx context.Context, destinationsDetails []operators.Destination, formationAssignment *model.FormationAssignment) error
}

// supportedDestinationsWithCertificate is a map of all destinations that as part of their creation a certificate resource is also created
var supportedDestinationsWithCertificate = map[string]bool{
	string(destinationcreatorpkg.AuthTypeSAMLAssertion):       true,
	string(destinationcreatorpkg.AuthTypeSAMLBearerAssertion): true,
	string(destinationcreatorpkg.AuthTypeClientCertificate):   true,
}

// Service consists of a service-level operations related to the destination entity
type Service struct {
	transact              persistence.Transactioner
	destinationRepo       destinationRepository
	tenantRepo            tenantRepository
	uidSvc                UIDService
	destinationCreatorSvc destinationCreatorService
}

// NewService creates a new Service
func NewService(
	transact persistence.Transactioner,
	destinationRepository destinationRepository,
	tenantRepository tenantRepository,
	uidSvc UIDService,
	destinationCreatorSvc destinationCreatorService,
) *Service {
	return &Service{
		transact:              transact,
		destinationRepo:       destinationRepository,
		tenantRepo:            tenantRepository,
		uidSvc:                uidSvc,
		destinationCreatorSvc: destinationCreatorSvc,
	}
}

// CreateDesignTimeDestinations is responsible to create so-called design time(destinationcreator.AuthTypeNoAuth) destination resource in the remote destination service as well as in our DB
func (s *Service) CreateDesignTimeDestinations(ctx context.Context, destinationsDetails []operators.Destination, formationAssignment *model.FormationAssignment) error {
	for _, destinationDetails := range destinationsDetails {
		if err := s.createDesignTimeDestinations(ctx, destinationDetails, formationAssignment); err != nil {
			return errors.Wrapf(err, "while creating design time destination with name: %q", destinationDetails.Name)
		}
	}

	return nil
}

func (s *Service) createDesignTimeDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *model.FormationAssignment) error {
	subaccountID, err := s.destinationCreatorSvc.ValidateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return err
	}
	destinationDetails.SubaccountID = subaccountID

	t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
	}

	tenantID := t.ID
	destinationFromDB, err := s.destinationRepo.GetDestinationByNameAndTenant(ctx, destinationDetails.Name, tenantID)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return err
		}
		log.C(ctx).Infof("No destination with name: %q and tenant ID: %q found in our DB, it will be created...", destinationDetails.Name, tenantID)
	}

	if destinationFromDB != nil && destinationFromDB.FormationAssignmentID != nil && *destinationFromDB.FormationAssignmentID != formationAssignment.ID {
		return errors.Errorf("Already have destination with name: %q and tenant ID: %q for assignment ID: %q. Could not have second destination with the same name and tenant ID but with different assignment ID: %q", destinationDetails.Name, tenantID, *destinationFromDB.FormationAssignmentID, formationAssignment.ID)
	}

	if err = s.destinationCreatorSvc.CreateDesignTimeDestinations(ctx, destinationDetails, formationAssignment, 0); err != nil {
		return err
	}

	destModel := &model.Destination{
		ID:                    s.uidSvc.Generate(),
		Name:                  destinationDetails.Name,
		Type:                  destinationDetails.Type,
		URL:                   destinationDetails.URL,
		Authentication:        destinationDetails.Authentication,
		SubaccountID:          t.ID,
		FormationAssignmentID: &formationAssignment.ID,
	}

	if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
		return errors.Wrapf(err, "while upserting design time destination with name: %q and assignment ID: %q in the DB", destinationDetails.Name, formationAssignment.ID)
	}

	return nil
}

// CreateBasicCredentialDestinations is responsible to create a basic destination resource in the remote destination service as well as in our DB
func (s *Service) CreateBasicCredentialDestinations(
	ctx context.Context,
	destinationsDetails []operators.Destination,
	basicAuthenticationCredentials operators.BasicAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) error {
	for _, destinationDetails := range destinationsDetails {
		if err := s.createBasicCredentialDestination(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs); err != nil {
			return errors.Wrapf(err, "while creating basic destination with name: %q", destinationDetails.Name)
		}
	}

	return nil
}

// CreateBasicCredentialDestination is responsible to create a basic destination resource in the remote destination service as well as in our DB
func (s *Service) createBasicCredentialDestination(
	ctx context.Context,
	destinationDetails operators.Destination,
	basicAuthenticationCredentials operators.BasicAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) error {
	subaccountID, err := s.destinationCreatorSvc.ValidateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return err
	}
	destinationDetails.SubaccountID = subaccountID

	t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
	}

	tenantID := t.ID
	destinationFromDB, err := s.destinationRepo.GetDestinationByNameAndTenant(ctx, destinationDetails.Name, tenantID)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return err
		}
		log.C(ctx).Infof("No destination with name: %q and tenant ID: %q found in our DB, it will be created...", destinationDetails.Name, tenantID)
	}

	if destinationFromDB != nil && destinationFromDB.FormationAssignmentID != nil && *destinationFromDB.FormationAssignmentID != formationAssignment.ID {
		return errors.Errorf("Already have destination with name: %q and tenant ID: %q for assignment ID: %q. Could not have second destination with the same name and tenant ID but with different assignment ID: %q", destinationDetails.Name, tenantID, *destinationFromDB.FormationAssignmentID, formationAssignment.ID)
	}

	if err = s.destinationCreatorSvc.CreateBasicCredentialDestinations(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, 0); err != nil {
		return err
	}

	basicReqBody, err := s.destinationCreatorSvc.PrepareBasicRequestBody(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs)
	if err != nil {
		return err
	}

	destModel := &model.Destination{
		ID:                    s.uidSvc.Generate(),
		Name:                  basicReqBody.Name,
		Type:                  string(basicReqBody.Type),
		URL:                   basicReqBody.URL,
		Authentication:        string(basicReqBody.AuthenticationType),
		SubaccountID:          t.ID,
		FormationAssignmentID: &formationAssignment.ID,
	}

	if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
		return errors.Wrapf(err, "while upserting basic destination with name: %q and assignment ID: %q in the DB", destinationDetails.Name, formationAssignment.ID)
	}

	return nil
}

// CreateSAMLAssertionDestination is responsible to create SAML assertion destination resource in the remote destination service as well as in our DB
func (s *Service) CreateSAMLAssertionDestination(
	ctx context.Context,
	destinationsDetails []operators.Destination,
	samlAssertionAuthCredentials *operators.SAMLAssertionAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) error {
	if err := s.destinationCreatorSvc.EnsureDestinationSubaccountIDsCorrectness(ctx, destinationsDetails, formationAssignment); err != nil {
		return errors.Wrap(err, "while ensuring the provided subaccount IDs in the destination details are correct")
	}

	for _, destinationDetails := range destinationsDetails {
		if err := s.createSAMLAssertionDestination(ctx, destinationDetails, samlAssertionAuthCredentials, formationAssignment, correlationIDs); err != nil {
			return errors.Wrapf(err, "while creating SAML Assertion destination with name: %q", destinationDetails.Name)
		}
	}

	return nil
}

// createSAMLAssertionDestination is responsible to create SAML assertion destination resource in the remote destination service as well as in our DB
func (s *Service) createSAMLAssertionDestination(
	ctx context.Context,
	destinationDetails operators.Destination,
	samlAssertionAuthCredentials *operators.SAMLAssertionAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) error {
	t, err := s.tenantRepo.GetByExternalTenant(ctx, destinationDetails.SubaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", destinationDetails.SubaccountID)
	}

	tenantID := t.ID
	destinationFromDB, err := s.destinationRepo.GetDestinationByNameAndTenant(ctx, destinationDetails.Name, tenantID)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return err
		}
		log.C(ctx).Infof("No destination with name: %q and tenant ID: %q found in our DB, it will be created...", destinationDetails.Name, tenantID)
	}

	if destinationFromDB != nil && destinationFromDB.FormationAssignmentID != nil && *destinationFromDB.FormationAssignmentID != formationAssignment.ID {
		return errors.Errorf("Already have destination with name: %q and tenant ID: %q for assignment ID: %q. Could not have second destination with the same name and tenant ID but with different assignment ID: %q", destinationDetails.Name, tenantID, *destinationFromDB.FormationAssignmentID, formationAssignment.ID)
	}

	if err = s.destinationCreatorSvc.CreateSAMLAssertionDestination(ctx, destinationDetails, samlAssertionAuthCredentials, formationAssignment, correlationIDs, 0); err != nil {
		return err
	}

	destModel := &model.Destination{
		ID:                    s.uidSvc.Generate(),
		Name:                  destinationDetails.Name,
		Type:                  string(destinationcreatorpkg.TypeHTTP),
		URL:                   samlAssertionAuthCredentials.URL,
		Authentication:        string(destinationcreatorpkg.AuthTypeSAMLAssertion),
		SubaccountID:          t.ID,
		FormationAssignmentID: &formationAssignment.ID,
	}

	if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
		return errors.Wrapf(err, "while upserting SAML Assertion destination with name: %q and assignment ID: %q in the DB", destinationDetails.Name, formationAssignment.ID)
	}

	return nil
}

// CreateClientCertificateAuthenticationDestination is responsible to create client certificate authentication destination resource in the remote destination service as well as in our DB
func (s *Service) CreateClientCertificateAuthenticationDestination(
	ctx context.Context,
	destinationsDetails []operators.Destination,
	clientCertAuthCredentials *operators.ClientCertAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) error {
	if err := s.destinationCreatorSvc.EnsureDestinationSubaccountIDsCorrectness(ctx, destinationsDetails, formationAssignment); err != nil {
		return errors.Wrap(err, "while ensuring the provided subaccount IDs in the destination details are correct")
	}

	for _, destinationDetails := range destinationsDetails {
		if err := s.createClientCertificateAuthenticationDestination(ctx, destinationDetails, clientCertAuthCredentials, formationAssignment, correlationIDs); err != nil {
			return errors.Wrapf(err, "while creating client certificate authentication destination with name: %q", destinationDetails.Name)
		}
	}

	return nil
}

func (s *Service) createClientCertificateAuthenticationDestination(
	ctx context.Context,
	destinationDetails operators.Destination,
	clientCertAuthCredentials *operators.ClientCertAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) error {
	t, err := s.tenantRepo.GetByExternalTenant(ctx, destinationDetails.SubaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", destinationDetails.SubaccountID)
	}

	tenantID := t.ID
	destinationFromDB, err := s.destinationRepo.GetDestinationByNameAndTenant(ctx, destinationDetails.Name, tenantID)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return err
		}
		log.C(ctx).Infof("No destination with name: %q and tenant ID: %q found in our DB, it will be created...", destinationDetails.Name, tenantID)
	}

	if destinationFromDB != nil && destinationFromDB.FormationAssignmentID != nil && *destinationFromDB.FormationAssignmentID != formationAssignment.ID {
		return errors.Errorf("Already have destination with name: %q and tenant ID: %q for assignment ID: %q. Could not have second destination with the same name and tenant ID but with different assignment ID: %q", destinationDetails.Name, tenantID, *destinationFromDB.FormationAssignmentID, formationAssignment.ID)
	}

	if err = s.destinationCreatorSvc.CreateClientCertificateDestination(ctx, destinationDetails, clientCertAuthCredentials, formationAssignment, correlationIDs, 0); err != nil {
		return err
	}

	destModel := &model.Destination{
		ID:                    s.uidSvc.Generate(),
		Name:                  destinationDetails.Name,
		Type:                  string(destinationcreatorpkg.TypeHTTP),
		URL:                   clientCertAuthCredentials.URL,
		Authentication:        string(destinationcreatorpkg.AuthTypeClientCertificate),
		SubaccountID:          t.ID,
		FormationAssignmentID: &formationAssignment.ID,
	}

	if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
		return errors.Wrapf(err, "while upserting SAML Assertion destination with name: %q and assignment ID: %q in the DB", destinationDetails.Name, formationAssignment.ID)
	}

	return nil
}

// DeleteDestinations is responsible to delete all types of destinations associated with the given `formationAssignment`
// from the DB as well as from the remote destination service
func (s *Service) DeleteDestinations(ctx context.Context, formationAssignment *model.FormationAssignment) error {
	externalDestSubaccountID, err := s.destinationCreatorSvc.GetConsumerTenant(ctx, formationAssignment)
	if err != nil {
		return err
	}

	formationAssignmentID := formationAssignment.ID

	t, err := s.tenantRepo.GetByExternalTenant(ctx, externalDestSubaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", externalDestSubaccountID)
	}

	destinations, err := s.destinationRepo.ListByTenantIDAndAssignmentID(ctx, t.ID, formationAssignmentID)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("There is/are %d destination(s) in the DB", len(destinations))
	if len(destinations) == 0 {
		return nil
	}

	for _, destination := range destinations {
		if supportedDestinationsWithCertificate[destination.Authentication] {
			certName, err := destinationcreator.GetDestinationCertificateName(ctx, destinationcreatorpkg.AuthType(destination.Authentication), formationAssignmentID)
			if err != nil {
				return errors.Wrapf(err, "while getting destination certificate name for destination auth type: %s", destination.Authentication)
			}
			if err = s.destinationCreatorSvc.DeleteCertificate(ctx, certName, externalDestSubaccountID, str.PtrStrToStr(destination.InstanceID), formationAssignment); err != nil {
				return errors.Wrapf(err, "while deleting destination certificate with name: %q", certName)
			}
		}

		if err := s.destinationCreatorSvc.DeleteDestination(ctx, destination.Name, externalDestSubaccountID, str.PtrStrToStr(destination.InstanceID), formationAssignment); err != nil {
			return err
		}

		if err := s.destinationRepo.DeleteByDestinationNameAndAssignmentID(ctx, destination.Name, formationAssignmentID, t.ID); err != nil {
			return errors.Wrapf(err, "while deleting destination(s) by name: %q, internal tenant ID: %q and assignment ID: %q from the DB", destination.Name, t.ID, formationAssignmentID)
		}
	}

	return nil
}
