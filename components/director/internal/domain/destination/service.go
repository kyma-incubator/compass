package destination

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=destinationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type destinationRepository interface {
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
	CreateSAMLAssertionDestination(ctx context.Context, destinationDetails operators.Destination, samlAuthCreds *operators.SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error
	DeleteDestination(ctx context.Context, destinationName, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) error
	DeleteCertificate(ctx context.Context, certificateName, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) error
	ValidateDestinationSubaccount(ctx context.Context, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) (string, error)
	PrepareBasicRequestBody(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) (*destinationcreator.BasicRequestBody, error)
	GetConsumerTenant(ctx context.Context, formationAssignment *model.FormationAssignment) (string, error)
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
func NewService(transact persistence.Transactioner, destinationRepository destinationRepository, tenantRepository tenantRepository, uidSvc UIDService, destinationCreatorSvc destinationCreatorService) *Service {
	return &Service{
		transact:              transact,
		destinationRepo:       destinationRepository,
		tenantRepo:            tenantRepository,
		uidSvc:                uidSvc,
		destinationCreatorSvc: destinationCreatorSvc,
	}
}

// CreateDesignTimeDestinations is responsible to create so-called design time(destinationcreator.AuthTypeNoAuth) destination resource in the remote destination service as well as in our DB
func (s *Service) CreateDesignTimeDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *model.FormationAssignment) error {
	if err := s.destinationCreatorSvc.CreateDesignTimeDestinations(ctx, destinationDetails, formationAssignment, 0); err != nil {
		return err
	}

	subaccountID, err := s.destinationCreatorSvc.ValidateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
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

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
			return errors.Wrapf(err, "while upserting basic destination with name: %q and assignment ID: %q in the DB", destinationDetails.Name, formationAssignment.ID)
		}
		return nil
	}); transactionErr != nil {
		return transactionErr
	}

	return nil
}

// CreateBasicCredentialDestinations is responsible to create a basic destination resource in the remote destination service as well as in our DB
func (s *Service) CreateBasicCredentialDestinations(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) error {
	if err := s.destinationCreatorSvc.CreateBasicCredentialDestinations(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, 0); err != nil {
		return err
	}

	subaccountID, err := s.destinationCreatorSvc.ValidateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	basicReqBody, err := s.destinationCreatorSvc.PrepareBasicRequestBody(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs)
	if err != nil {
		return err
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
		if err != nil {
			return errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
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
	}); transactionErr != nil {
		return transactionErr
	}

	return nil
}

// CreateSAMLAssertionDestination is responsible to create SAML assertion destination resource in the remote destination service as well as in our DB
func (s *Service) CreateSAMLAssertionDestination(ctx context.Context, destinationDetails operators.Destination, samlAuthCreds *operators.SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) error {
	if err := s.destinationCreatorSvc.CreateSAMLAssertionDestination(ctx, destinationDetails, samlAuthCreds, formationAssignment, correlationIDs, 0); err != nil {
		return err
	}

	subaccountID, err := s.destinationCreatorSvc.ValidateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
	}

	destModel := &model.Destination{
		ID:                    s.uidSvc.Generate(),
		Name:                  destinationDetails.Name,
		Type:                  string(destinationcreator.TypeHTTP),
		URL:                   samlAuthCreds.URL,
		Authentication:        string(destinationcreator.AuthTypeSAMLAssertion),
		SubaccountID:          t.ID,
		FormationAssignmentID: &formationAssignment.ID,
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
			return errors.Wrapf(err, "while upserting basic destination with name: %q and assignment ID: %q in the DB", destinationDetails.Name, formationAssignment.ID)
		}
		return nil
	}); transactionErr != nil {
		return transactionErr
	}

	return nil
}

// DeleteDestinations is responsible to delete all type of destinations associated with the given `formationAssignment` from the DB as well as from the remote destination service
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
		if destination.Authentication == string(destinationcreator.AuthTypeSAMLAssertion) {
			if err := s.destinationCreatorSvc.DeleteCertificate(ctx, destination.Name, externalDestSubaccountID, formationAssignment); err != nil {
				return errors.Wrapf(err, "while deleting SAML assertion certificate with name: %q", destination.Name)
			}
		}
		if err := s.destinationCreatorSvc.DeleteDestination(ctx, destination.Name, externalDestSubaccountID, formationAssignment); err != nil {
			return err
		}

		if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
			if err := s.destinationRepo.DeleteByDestinationNameAndAssignmentID(ctx, destination.Name, formationAssignmentID, t.ID); err != nil {
				return errors.Wrapf(err, "while deleting destination(s) by name: %q, internal tenant ID: %q and assignment ID: %q from the DB", destination.Name, t.ID, formationAssignmentID)
			}
			return nil
		}); transactionErr != nil {
			return transactionErr
		}
	}

	return nil
}

func (s *Service) transaction(ctx context.Context, dbCall func(ctxWithTransact context.Context) error) error {
	tx, err := s.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to begin DB transaction")
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = dbCall(ctx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("Failed to commit database transaction")
		return err
	}
	return nil
}
