package consumer

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type ConsumerType string

const (
	Runtime           ConsumerType = "Runtime"
	Application       ConsumerType = "Application"
	IntegrationSystem ConsumerType = "Integration System"
	User              ConsumerType = "Static User"
)

type Consumer struct {
	ConsumerID string
	ConsumerType
}

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
