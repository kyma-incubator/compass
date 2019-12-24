package graphql

import (
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/pkg/errors"
)

func ToViewer(cons consumer.Consumer) (*Viewer, error) {
	switch cons.ConsumerType {
	case consumer.Runtime:
		return &Viewer{ID: cons.ConsumerID, Type: ViewerTypeRuntime}, nil
	case consumer.Application:
		return &Viewer{ID: cons.ConsumerID, Type: ViewerTypeApplication}, nil
	case consumer.IntegrationSystem:
		return &Viewer{ID: cons.ConsumerID, Type: ViewerTypeIntegrationSystem}, nil
	case consumer.User:
		return &Viewer{ID: cons.ConsumerID, Type: ViewerTypeUser}, nil
	}

	return nil, errors.New("Viewer not exist")

}
