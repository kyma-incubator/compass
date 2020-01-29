package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

// Unbind deletes an existing service binding
//   DELETE /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	b.dumper.Dump("Unbind instanceID:", instanceID)
	b.dumper.Dump("Unbind details:", details)
	b.dumper.Dump("Unbind asyncAllowed:", asyncAllowed)

	unbind := domain.UnbindSpec{}
	return unbind, nil
}
