package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const legacyServicesLabelKey = "compass/legacy-services"

type LegacyServiceReference struct {
	ID         string  `json:"id"`
	APIDefID   *string `json:"apiDefID"`
	EventDefID *string `json:"eventDefID"`
}

type labeler struct{}

func NewAppLabeler() *labeler {
	return &labeler{}
}

func (l *labeler) WriteServiceReference(appDetails graphql.ApplicationExt, serviceReference LegacyServiceReference) (graphql.LabelInput, error) {
	services, err := l.readLabel(appDetails)
	if err != nil {
		return graphql.LabelInput{}, err
	}

	services[serviceReference.ID] = serviceReference

	return l.writeLabel(services)
}

func (l *labeler) ReadServiceReference(appDetails graphql.ApplicationExt, serviceID string) (LegacyServiceReference, error) {
	services, err := l.readLabel(appDetails)
	if err != nil {
		return LegacyServiceReference{}, err
	}

	service, exists := services[serviceID]
	if !exists {
		return LegacyServiceReference{}, apperrors.NotFound("service with ID '%s' not found", serviceID)
	}

	return service, nil
}

func (l *labeler) DeleteServiceReference(appDetails graphql.ApplicationExt, serviceID string) (graphql.LabelInput, error) {
	services, err := l.readLabel(appDetails)
	if err != nil {
		return graphql.LabelInput{}, err
	}

	delete(services, serviceID)

	return l.writeLabel(services)
}

func (l *labeler) ReadService(appDetails graphql.ApplicationExt, serviceID string) (model.GraphQLServiceDetails, error) {
	serviceRef, err := l.ReadServiceReference(appDetails, serviceID)
	if err != nil {
		return model.GraphQLServiceDetails{}, err
	}

	var outputApi *graphql.APIDefinitionExt
	if serviceRef.APIDefID != nil {
		for _, api := range appDetails.APIDefinitions.Data {
			if api != nil && api.ID == *serviceRef.APIDefID {
				outputApi = api
				break
			}
		}
	}

	var outputEvent *graphql.EventAPIDefinitionExt
	if serviceRef.EventDefID != nil {
		for _, event := range appDetails.EventDefinitions.Data {
			if event != nil && event.ID == *serviceRef.EventDefID {
				outputEvent = event
				break
			}
		}
	}

	return model.GraphQLServiceDetails{
		ID:    serviceRef.ID,
		API:   outputApi,
		Event: outputEvent,
	}, nil
}

func (l *labeler) ListServices(appDetails graphql.ApplicationExt) ([]model.GraphQLServiceDetails, error) {
	services, err := l.readLabel(appDetails)
	if err != nil {
		return nil, err
	}
	var serviceReferences []model.GraphQLServiceDetails
	for serviceIDKey, _ := range services {
		value, err := l.ReadService(appDetails, serviceIDKey)
		if err != nil {
			return nil, err
		}
		serviceReferences = append(serviceReferences, value)
	}
	return serviceReferences, nil
}

func (l *labeler) readLabel(appDetails graphql.ApplicationExt) (map[string]LegacyServiceReference, error) {
	value := appDetails.Labels[legacyServicesLabelKey]
	if value == nil {
		value = "{}"
	}

	strValue, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid type: expected: string; actual: %T", value)
	}

	var services map[string]LegacyServiceReference

	err := json.Unmarshal([]byte(strValue), &services)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling JSON value")
	}

	return services, nil
}

func (l *labeler) writeLabel(services map[string]LegacyServiceReference) (graphql.LabelInput, error) {
	marshalledServices, err := json.Marshal(services)
	if err != nil {
		return graphql.LabelInput{}, errors.Wrap(err, "while marshalling JSON value")
	}

	return graphql.LabelInput{
		Key:   legacyServicesLabelKey,
		Value: strconv.Quote(string(marshalledServices)),
	}, nil
}
