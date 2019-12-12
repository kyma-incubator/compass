package tenantmapping

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

type ConsumerType string

const (
	RUNTIME            ConsumerType = "Runtime"
	APPLICATION        ConsumerType = "Application"
	INTEGRATION_SYSTEM ConsumerType = "Integration System"
	USER               ConsumerType = "Static User"
)

type Consumer struct {
	ConsumerID string
	ConsumerType
}

func MapSystemAuthToConsumerType(refObj model.SystemAuthReferenceObjectType) (ConsumerType, error) {
	switch refObj {
	case model.ApplicationReference:
		return APPLICATION, nil
	case model.RuntimeReference:
		return RUNTIME, nil
	case model.IntegrationSystemReference:
		return INTEGRATION_SYSTEM, nil
	}
	return "", errors.New("unknown reference object type")
}
