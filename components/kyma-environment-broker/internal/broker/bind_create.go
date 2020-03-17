package broker

import (
	"context"
	"errors"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
)

type BindEndpoint struct {
	log logrus.FieldLogger
}

func NewBind(log logrus.FieldLogger) *BindEndpoint {
	return &BindEndpoint{log: log.WithField("service", "BindEndpoint")}
}

// Bind creates a new service binding
//   PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *BindEndpoint) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	b.log.Infof("Bind instanceID:", instanceID)
	b.log.Infof("Bind parameters: %s", string(details.RawParameters))
	b.log.Infof("Bind context: %s", string(details.RawContext))
	b.log.Infof("Bind asyncAllowed:", asyncAllowed)

	return domain.Binding{}, errors.New("not supported")
}
