package process

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

type StepProcessed struct {
	StepName string
	Duration time.Duration
	When     time.Duration
	Error    error
}

type ProvisioningStepProcessed struct {
	StepProcessed
	OldOperation internal.ProvisioningOperation
	Operation    internal.ProvisioningOperation
}

type DeprovisioningStepProcessed struct {
	StepProcessed
	OldOperation internal.DeprovisioningOperation
	Operation    internal.DeprovisioningOperation
}
