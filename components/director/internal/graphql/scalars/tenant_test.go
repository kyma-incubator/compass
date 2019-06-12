package scalars

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenant_UnmarshalGQL(t *testing.T) {
	//given
	var tenant Tenant
	fixTenant := "tenant1"
	expectedTenant := Tenant("tenant1")

	//when
	err := tenant.UnmarshalGQL(fixTenant)

	//then
	require.NoError(t, err)
	assert.Equal(t, expectedTenant, tenant)
}

func TestTenant_MarshalGQL(t *testing.T) {
	//given
	fixTenant := Tenant("tenant1")
	expectedTenant := `{"tenant":"tenant1"}`
	buf := bytes.Buffer{}
	//when
	fixTenant.MarshalGQL(&buf)

	//then
	assert.Equal(t, expectedTenant, buf.String())
}
