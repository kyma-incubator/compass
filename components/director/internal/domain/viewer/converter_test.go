package viewer_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/viewer"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToViewer(t *testing.T) {
	id := uuid.New().String()
	testCases := []struct {
		Name        string
		Input       consumer.Consumer
		Expected    graphql.Viewer
		ExpectedErr bool
	}{
		{
			Name:     "Convert to Runtime",
			Input:    consumer.Consumer{ConsumerID: id, ConsumerType: consumer.Runtime},
			Expected: graphql.Viewer{ID: id, Type: graphql.ViewerTypeRuntime},
		},
		{
			Name:     "Convert To Application",
			Input:    consumer.Consumer{ConsumerID: id, ConsumerType: consumer.Application},
			Expected: graphql.Viewer{ID: id, Type: graphql.ViewerTypeApplication},
		},
		{
			Name:     "Convert To Integration System",
			Input:    consumer.Consumer{ConsumerID: id, ConsumerType: consumer.IntegrationSystem},
			Expected: graphql.Viewer{ID: id, Type: graphql.ViewerTypeIntegrationSystem},
		},
		{
			Name:     "Convert To User",
			Input:    consumer.Consumer{ConsumerID: id, ConsumerType: consumer.User},
			Expected: graphql.Viewer{ID: id, Type: graphql.ViewerTypeUser},
		},
		{
			Name:        "Error while converting",
			Input:       consumer.Consumer{ConsumerType: "Janusz"},
			ExpectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// WHEN
			viewer, err := viewer.ToViewer(tc.Input)
			// THEN
			if tc.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NotNil(t, viewer)
				assert.Equal(t, tc.Expected, *viewer)
			}
		})
	}
}
