package broker

import (
	"context"
	"errors"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
)

type UnbindEndpoint struct {
	log logrus.FieldLogger
}

func NewUnbind(log logrus.FieldLogger) *UnbindEndpoint {
	return &UnbindEndpoint{log: log.WithField("service", "UnbindEndpoint")}
}

// Unbind deletes an existing service binding
//   DELETE /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *UnbindEndpoint) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	b.log.Infof("Unbind instanceID: %s", instanceID)
	b.log.Infof("Unbind details: %+v", details)
	b.log.Infof("Unbind asyncAllowed: %v", asyncAllowed)

	return domain.UnbindSpec{}, errors.New("not supported")
}
