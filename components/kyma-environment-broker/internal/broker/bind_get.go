package broker

import (
	"context"
	"errors"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
)

type GetBindingEndpoint struct {
	log logrus.FieldLogger
}

func NewGetBinding(log logrus.FieldLogger) *GetBindingEndpoint {
	return &GetBindingEndpoint{log: log.WithField("service", "GetBindingEndpoint")}
}

// GetBinding fetches an existing service binding
//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *GetBindingEndpoint) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	b.log.Infof("GetBinding instanceID: %s", instanceID)
	b.log.Infof("GetBinding bindingID: %s", bindingID)

	return domain.GetBindingSpec{}, errors.New("not supported")
}
