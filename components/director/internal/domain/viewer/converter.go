package viewer

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// ToViewer missing godoc
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

	return nil, apperrors.NewInternalError("viewer does not exist")
}
