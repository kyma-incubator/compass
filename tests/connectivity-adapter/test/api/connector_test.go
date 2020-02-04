package api

import (
	"testing"

	"github.com/kyma-incubator/compass/tests/connectivity-adapter/test/testkit"
	"github.com/stretchr/testify/require"
)

func TestConnector(t *testing.T) {
	_, err := testkit.ReadConfiguration()
	require.NoError(t, err)
}
