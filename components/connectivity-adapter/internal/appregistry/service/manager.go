package service

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=DirectorClient -output=automock -outpkg=automock -case=underscore
type DirectorClient interface {
	CreateAPIDefinition(appID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
	CreateEventDefinition(appID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)

	SetApplicationLabel(appID string, label graphql.LabelInput) error
	GetApplicationsByNameRequest(appName string) *gcli.Request
	//DeleteAPIDefinition(apiID string) (string, error)
	//DeleteEventDefinition(eventID string) (string, error)
}

type AppLabeler interface {
	WriteService(appDetails graphql.ApplicationExt, serviceReference LegacyServiceReference) (graphql.LabelInput, error)
	ReadService(appDetails graphql.ApplicationExt, serviceID string) GraphQLServiceDetails
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

	var apiID, eventID *string
	if serviceDetails.API != nil {
		id, err := s.directorClient.CreateAPIDefinition(appID, *serviceDetails.API)
		if err != nil {
			return errors.Wrap(err, "while creating API Definition")
		}

		apiID = &id
	}

	if serviceDetails.Event != nil {
		id, err := s.directorClient.CreateEventDefinition(appID, *serviceDetails.Event)
		if err != nil {
			return errors.Wrap(err, "while creating Event API Definition")
		}

		eventID = &id
	}

	label, err := s.appLabeler.WriteService(s.appDetails, LegacyServiceReference{
		ID:         serviceDetails.ID,
		APIDefID:   apiID,
		EventDefID: eventID,
	})
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

func (s *serviceManager) GetFromApplicationDetails(serviceID string) (model.GraphQLServiceDetails, error) {
	panic("implement me")
}

func (s *serviceManager) ListFromApplicationDetails() ([]model.GraphQLServiceDetails, error) {
	panic("implement me")
}

func (s *serviceManager) Update(serviceDetails model.GraphQLServiceDetailsInput) error {
	panic("implement me")
}

func (s *serviceManager) Delete(serviceID string) error {
	serviceDetails, err := s.GetFromApplicationDetails(serviceID)
	if err != nil {
		return err
	}

	if serviceDetails.API != nil {
		// TODO: Delete API
	}

	if serviceDetails.Event != nil {
		// TODO: Delete Event
	}

	// TODO: Remove service from App label

	return nil
}
