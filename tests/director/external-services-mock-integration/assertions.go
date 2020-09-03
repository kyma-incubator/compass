package external_services_mock_integration

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func assertSpecInBundleNotNil(t *testing.T, bundle graphql.BundleExt) {
	assert.True(t, len(bundle.APIDefinitions.Data) > 0)
	assert.NotNil(t, bundle.APIDefinitions.Data[0])
	assert.NotNil(t, bundle.APIDefinitions.Data[0].Spec)
	assert.NotNil(t, bundle.APIDefinitions.Data[0].Spec.Data)
}
