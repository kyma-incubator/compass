package broker

import (
	"context"
	"errors"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type BindEndpoint struct {
	dumper StructDumper
}

func NewBind(dumper StructDumper) *BindEndpoint {
	return &BindEndpoint{dumper: dumper}
}

// Bind creates a new service binding
//   PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *BindEndpoint) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	b.dumper.Dump("Bind instanceID:", instanceID)
	b.dumper.Dump("Bind details:", details)
	b.dumper.Dump("Bind asyncAllowed:", asyncAllowed)

	return domain.Binding{}, errors.New("not supported")
}
