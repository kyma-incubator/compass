package azure

import (
	"context"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
)

type EventHub struct {
	HyperscalerProvider azure.HyperscalerProvider
	AccountProvider     hyperscaler.AccountProvider
	Context             context.Context
}
