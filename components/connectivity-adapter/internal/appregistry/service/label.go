package service

import (
	"encoding/json"
	"fmt"
	"strconv"

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

func (l *labeler) WriteService(appDetails graphql.ApplicationExt, serviceReference LegacyServiceReference) (graphql.LabelInput, error) {
	value := appDetails.Labels[legacyServicesLabelKey]
	if value == nil {
		value = "{}"
	}

	strValue, ok := value.(string)
	if !ok {
		return graphql.LabelInput{}, fmt.Errorf("invalid type: expected: string; actual: %T", value)
	}

	var services map[string]LegacyServiceReference

	err := json.Unmarshal([]byte(strValue), &services)
	if err != nil {
		return graphql.LabelInput{}, errors.Wrap(err, "while unmarshalling JSON value")
	}

	services[serviceReference.ID] = serviceReference

	marshalledServices, err := json.Marshal(services)
	if err != nil {
		return graphql.LabelInput{}, errors.Wrap(err, "while marshalling JSON value")
	}

	return graphql.LabelInput{
		Key:   legacyServicesLabelKey,
		Value: strconv.Quote(string(marshalledServices)),
	}, nil
}

func (l *labeler) ReadService(appDetails graphql.ApplicationExt, serviceID string) GraphQLServiceDetails {
	panic("implement me")
}
