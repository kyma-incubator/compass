package service

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=DirectorClient -output=automock -outpkg=automock -case=underscore
type DirectorClient interface {
	CreateAPIDefinition(appID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
	CreateEventDefinition(appID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)

	SetApplicationLabel(appID string, label graphql.LabelInput) error
	DeleteAPIDefinition(apiID string) error
	DeleteEventDefinition(eventID string) error
}

//go:generate mockery -name=AppLabeler -output=automock -outpkg=automock -case=underscore
type AppLabeler interface {
	WriteServiceReference(appDetails graphql.ApplicationExt, serviceReference LegacyServiceReference) (graphql.LabelInput, error)
	DeleteServiceReference(appDetails graphql.ApplicationExt, serviceID string) (graphql.LabelInput, error)
	ReadServiceReference(appDetails graphql.ApplicationExt, serviceID string) (LegacyServiceReference, error)
	ReadService(appDetails graphql.ApplicationExt, serviceID string) (model.GraphQLServiceDetails, error)
	ListServices(appDetails graphql.ApplicationExt) ([]model.GraphQLServiceDetails, error)
}

type serviceManager struct {
	appDetails     graphql.ApplicationExt
	directorClient DirectorClient
	appLabeler     AppLabeler
}

func NewServiceManager(directorCli DirectorClient, appLabeler AppLabeler, appDetails graphql.ApplicationExt) (*serviceManager, error) {
	return &serviceManager{
		appDetails:     appDetails,
		directorClient: directorCli,
		appLabeler:     appLabeler,
	}, nil
}

func (s *serviceManager) Create(serviceDetails model.GraphQLServiceDetailsInput) error {
	appID := s.appDetails.ID

	serviceRef, err := s.createAPIandEventDefinitions(appID, serviceDetails)
	if err != nil {
		return err
	}

	err = s.setAppLabelWithServiceRef(appID, serviceRef)
	if err != nil {
		// TODO: revert creating API and EventAPI definitions
		return err
	}

	return nil
}

func (s *serviceManager) GetFromApplicationDetails(serviceID string) (model.GraphQLServiceDetails, error) {
	return s.appLabeler.ReadService(s.appDetails, serviceID)
}

func (s *serviceManager) ListFromApplicationDetails() ([]model.GraphQLServiceDetails, error) {
	return s.appLabeler.ListServices(s.appDetails)
}

func (s *serviceManager) Update(serviceDetails model.GraphQLServiceDetailsInput) error {
	appID := s.appDetails.ID

	serviceRef, err := s.appLabeler.ReadServiceReference(s.appDetails, serviceDetails.ID)
	if err != nil {
		return err
	}

	err = s.deleteAPIandEventDefinitions(serviceRef)
	if err != nil {
		return err
	}

	newServiceRef, err := s.createAPIandEventDefinitions(appID, serviceDetails)
	if err != nil {
		return err
	}

	err = s.setAppLabelWithServiceRef(appID, newServiceRef)
	if err != nil {
		// TODO: revert creating API and EventAPI definitions?
		return err
	}

	return nil
}

func (s *serviceManager) Delete(serviceID string) error {
	serviceRef, err := s.appLabeler.ReadServiceReference(s.appDetails, serviceID)
	if err != nil {
		return err
	}

	err = s.deleteAPIandEventDefinitions(serviceRef)
	if err != nil {
		return err
	}

	label, err := s.appLabeler.DeleteServiceReference(s.appDetails, serviceRef.ID)
	if err != nil {
		return errors.Wrap(err, "while writing Application label")

		// TODO: Should we somehow restore deleted API and Event Definitions? ( ಠ_ಠ)
	}

	err = s.directorClient.SetApplicationLabel(s.appDetails.Application.ID, label)
	if err != nil {
		return errors.Wrap(err, "while setting Application label")
		// TODO: revert creating API and EventAPI definitions
	}

	return nil
}

func (s *serviceManager) createAPIandEventDefinitions(appID string, serviceDetails model.GraphQLServiceDetailsInput) (LegacyServiceReference, error) {
	var apiID, eventID *string
	if serviceDetails.API != nil {
		id, err := s.directorClient.CreateAPIDefinition(appID, *serviceDetails.API)
		if err != nil {
			return LegacyServiceReference{}, errors.Wrap(err, "while creating API Definition")
		}

		apiID = &id
	}

	if serviceDetails.Event != nil {
		id, err := s.directorClient.CreateEventDefinition(appID, *serviceDetails.Event)
		if err != nil {
			return LegacyServiceReference{}, errors.Wrap(err, "while creating Event API Definition")
		}

		eventID = &id
	}

	return LegacyServiceReference{
		ID:         serviceDetails.ID,
		APIDefID:   apiID,
		EventDefID: eventID,
	}, nil
}

func (s *serviceManager) deleteAPIandEventDefinitions(serviceRef LegacyServiceReference) error {
	if serviceRef.APIDefID != nil {
		err := s.directorClient.DeleteAPIDefinition(*serviceRef.APIDefID)
		if err != nil {
			return errors.Wrap(err, "while deleting API Definition")
		}
	}

	if serviceRef.EventDefID != nil {
		err := s.directorClient.DeleteAPIDefinition(*serviceRef.EventDefID)
		if err != nil {
			return errors.Wrap(err, "while deleting Event Definition")
		}
	}

	return nil
}

func (s *serviceManager) setAppLabelWithServiceRef(appID string, serviceRef LegacyServiceReference) error {
	label, err := s.appLabeler.WriteServiceReference(s.appDetails, serviceRef)
	if err != nil {
		return errors.Wrap(err, "while writing Application label")
		// TODO: revert creating API and EventAPI definitions
	}

	err = s.directorClient.SetApplicationLabel(appID, label)
	if err != nil {
		return errors.Wrap(err, "while setting Application label")
		// TODO: revert creating API and EventAPI definitions
	}

	return nil
}
