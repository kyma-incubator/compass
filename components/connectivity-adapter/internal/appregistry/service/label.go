package service

import (
	"fmt"

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

func (l *labeler) WriteService(appDetails graphql.ApplicationExt, serviceReference LegacyServiceReference) (graphql.LabelInput, error) {
	value := appDetails.Labels[legacyServicesLabelKey]

	if value == nil {
		value = make(map[string]LegacyServiceReference)
	}

	services, ok := value.(map[string]interface{})
	if !ok {
		return graphql.LabelInput{}, fmt.Errorf("invalid type: expected: map[string]LegacyServiceReference; actual: %T", value)
	}

	services[serviceReference.ID] = serviceReference

	return graphql.LabelInput{
		Key:   legacyServicesLabelKey,
		Value: services,
	}, nil
}

func (l *labeler) ReadService(appDetails graphql.ApplicationExt, serviceID string) GraphQLServiceDetails {
	panic("implement me")
}
