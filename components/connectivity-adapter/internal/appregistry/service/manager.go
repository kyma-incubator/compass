package service

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=GraphQLRequester -output=automock -outpkg=automock -case=underscore
type GraphQLRequester interface {
	CreateAPIDefinition(appID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error)
	CreateEventDefinition(appID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error)

	SetApplicationLabel(appID string, label graphql.LabelInput) error

	//DeleteAPIDefinition(apiID string) (string, error)
	//DeleteEventDefinition(eventID string) (string, error)
}

type AppLabeler interface {
	WriteService(appDetails graphql.ApplicationExt, serviceReference LegacyServiceReference) (graphql.LabelInput, error)
	ReadService(appDetails graphql.ApplicationExt, serviceID string) GraphQLServiceDetails
}

type serviceManager struct {
	appDetails   graphql.ApplicationExt
	gqlRequester GraphQLRequester
	appLabeler   AppLabeler
}

func NewServiceManager(gqlRequester GraphQLRequester, appLabeler AppLabeler, appDetails graphql.ApplicationExt) (*serviceManager, error) {
	return &serviceManager{
		appDetails:   appDetails,
		gqlRequester: gqlRequester,
		appLabeler:   appLabeler,
	}, nil
}

func (s *serviceManager) Create(serviceDetails GraphQLServiceDetailsInput) error {
	appID := s.appDetails.ID

	var apiID, eventID *string
	if serviceDetails.API != nil {
		id, err := s.gqlRequester.CreateAPIDefinition(appID, *serviceDetails.API)
		if err != nil {
			return errors.Wrap(err, "while creating API Definition")
		}

		apiID = &id
	}

	if serviceDetails.Event != nil {
		id, err := s.gqlRequester.CreateEventDefinition(appID, *serviceDetails.Event)
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

	err = s.gqlRequester.SetApplicationLabel(appID, label)
	if err != nil {
		return errors.Wrap(err, "while setting Application label")
		// TODO: revert creating API and EventAPI definitions
	}

	return nil
}

func (s *serviceManager) GetFromApplicationDetails(serviceID string) (GraphQLServiceDetails, error) {
	panic("implement me")
}

func (s *serviceManager) ListFromApplicationDetails() ([]GraphQLServiceDetails, error) {
	panic("implement me")
}

func (s *serviceManager) Update(serviceDetails GraphQLServiceDetailsInput) error {
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
