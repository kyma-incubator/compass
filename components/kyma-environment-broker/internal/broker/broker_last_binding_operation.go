package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

// LastBindingOperation fetches last operation state for a service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation
func (b *KymaEnvBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	b.dumper.Dump("LastBindingOperation instanceID:", instanceID)
	b.dumper.Dump("LastBindingOperation bindingID:", bindingID)
	b.dumper.Dump("LastBindingOperation details:", details)

	op := domain.LastOperation{}
	return op, nil
}
