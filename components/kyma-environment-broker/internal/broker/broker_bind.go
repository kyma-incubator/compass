package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

// Bind creates a new service binding
//   PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *KymaEnvBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	b.dumper.Dump("Bind instanceID:", instanceID)
	b.dumper.Dump("Bind details:", details)
	b.dumper.Dump("Bind asyncAllowed:", asyncAllowed)

	binding := domain.Binding{
		Credentials: map[string]interface{}{
			"host":     "test",
			"port":     "1234",
			"password": "nimda123",
		},
	}
	return binding, nil
}
