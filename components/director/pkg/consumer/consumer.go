package consumer

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
)

// ConsumerType missing godoc
type ConsumerType string

const (
	// Runtime missing godoc
	Runtime ConsumerType = "Runtime"
	// Application missing godoc
	Application ConsumerType = "Application"
	// IntegrationSystem missing godoc
	IntegrationSystem ConsumerType = "Integration System"
	// User missing godoc
	User ConsumerType = "Static User"
)

// Consumer missing godoc
type Consumer struct {
	ConsumerID   string `json:"ConsumerID"`
	ConsumerType `json:"ConsumerType"`
	Flow         oathkeeper.AuthFlow `json:"Flow"`
	OnBehalfOf   string              `json:"onBehalfOf"`
}

// MapSystemAuthToConsumerType missing godoc
func MapSystemAuthToConsumerType(refObj model.SystemAuthReferenceObjectType) (ConsumerType, error) {
	switch refObj {
	case model.ApplicationReference:
		return Application, nil
	case model.RuntimeReference:
		return Runtime, nil
	case model.IntegrationSystemReference:
		return IntegrationSystem, nil
	}
	return "", apperrors.NewInternalError("unknown reference object type")
}
