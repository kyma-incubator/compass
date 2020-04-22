package provisioning

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuditLog_FileDoesNotExist(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	svc := NewAuditLogOverridesStep(memoryStorage.Operations())

	operation := internal.ProvisioningOperation{}

	// when
	_, _, err := svc.Run(operation, NewLogDummy())
	//then
	require.Error(t, err)
	require.EqualError(t, err, "open audit-log-config: no such file or directory")

}

