package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

// GetBinding fetches an existing service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	b.dumper.Dump("GetBinding instanceID:", instanceID)
	b.dumper.Dump("GetBinding bindingID:", bindingID)

	spec := domain.GetBindingSpec{}
	return spec, nil
}
