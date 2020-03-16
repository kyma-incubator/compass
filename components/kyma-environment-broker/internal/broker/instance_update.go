package broker

import (
	"context"
	"errors"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
)

type UpdateEndpoint struct {
	log logrus.FieldLogger
}

func NewUpdate(log logrus.FieldLogger) *UpdateEndpoint {
	return &UpdateEndpoint{log: log.WithField("service", "UpdateEndpoint")}
}

// Update modifies an existing service instance
//  PATCH /v2/service_instances/{instance_id}
func (b *UpdateEndpoint) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	b.log.Infof("Update instanceID: %s", instanceID)
	b.log.Infof("Update details: %+v", details)
	b.log.Infof("Update asyncAllowed: %v", asyncAllowed)

	return domain.UpdateServiceSpec{}, errors.New("not supported")
}
