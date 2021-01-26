package external_services_mock_integration

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func assertSpecInBundleNotNil(t *testing.T, bndl graphql.BundleExt) {
	assert.True(t, len(bndl.APIDefinitions.Data) > 0)
	assert.NotNil(t, bndl.APIDefinitions.Data[0])
	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec)
	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec.Data)
}
