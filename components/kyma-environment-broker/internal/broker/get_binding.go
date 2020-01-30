package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type GetBindingEndpoint struct {
	dumper StructDumper
}

func NewGetBinding(dumper StructDumper) *GetBindingEndpoint {
	return &GetBindingEndpoint{dumper: dumper}
}

// GetBinding fetches an existing service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *GetBindingEndpoint) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	b.dumper.Dump("GetBinding instanceID:", instanceID)
	b.dumper.Dump("GetBinding bindingID:", bindingID)

	spec := domain.GetBindingSpec{}
	return spec, nil
}
