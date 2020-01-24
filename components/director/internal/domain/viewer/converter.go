package viewer

import (
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

func ToViewer(cons consumer.Consumer) (*graphql.Viewer, error) {
	switch cons.ConsumerType {
	case consumer.Runtime:
		return &graphql.Viewer{ID: cons.ConsumerID, Type: graphql.ViewerTypeRuntime}, nil
	case consumer.Application:
		return &graphql.Viewer{ID: cons.ConsumerID, Type: graphql.ViewerTypeApplication}, nil
	case consumer.IntegrationSystem:
		return &graphql.Viewer{ID: cons.ConsumerID, Type: graphql.ViewerTypeIntegrationSystem}, nil
	case consumer.User:
		return &graphql.Viewer{ID: cons.ConsumerID, Type: graphql.ViewerTypeUser}, nil
	}

	return nil, errors.New("viewer does not exist")

}
