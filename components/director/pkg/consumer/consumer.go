package consumer

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"
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
func MapSystemAuthToConsumerType(refObj systemauth.SystemAuthReferenceObjectType) (ConsumerType, error) {
	switch refObj {
	case systemauth.ApplicationReference:
		return Application, nil
	case systemauth.RuntimeReference:
		return Runtime, nil
	case systemauth.IntegrationSystemReference:
		return IntegrationSystem, nil
	}
	return "", apperrors.NewInternalError("unknown reference object type")
}
