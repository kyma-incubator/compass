package provisioning

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
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

func TestAuditLog_HappyPath(t *testing.T) {
	// given
	mm := afero.NewMemMapFs()
	err := mm.MkdirAll("audit-log-config", 0755)
	if err != nil {
		t.Fatalf("Unable to create file: audit-log-config!!: %v", err)
	}
	fileData := `
	- east:
		host: host
		http-user: aaaa
		http-pwd: aaaa
	
	- west:
		host: host
		http-user: bbbb
		http-pwd: bbbb
	`
	err = afero.WriteFile(mm, "audit-log-config", []byte(fileData), 0755)
	if err != nil {
		t.Fatalf("Unable to write contents to file: audit-log-config!!: %v", err)
	}

	memoryStorage := storage.NewMemoryStorage()
	svc := NewAuditLogOverridesStep(memoryStorage.Operations())

	operation := internal.ProvisioningOperation{}

	// when
	_, _, err = svc.Run(operation, NewLogDummy())
	//then
	require.Error(t, err)
	t.Log(err)


}
