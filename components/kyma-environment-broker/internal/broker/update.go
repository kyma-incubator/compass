package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type UpdateEndpoint struct {
	dumper StructDumper
}

func NewUpdate(dumper StructDumper) *UpdateEndpoint {
	return &UpdateEndpoint{dumper: dumper}
}

// Update modifies an existing service instance
//  PATCH /v2/service_instances/{instance_id}
func (b *UpdateEndpoint) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	b.dumper.Dump("Update instanceID:", instanceID)
	b.dumper.Dump("Update details:", details)
	b.dumper.Dump("Update asyncAllowed:", asyncAllowed)

	return domain.UpdateServiceSpec{}, nil
}
