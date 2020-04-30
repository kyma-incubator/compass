package provisioning

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/auditlog"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLog_ScriptFileDoesNotExist(t *testing.T) {
	// given
	mm := afero.NewMemMapFs()

	repo := storage.NewMemoryStorage().Operations()
	cfg := auditlog.Config{
		URL:      "host1",
		User:     "aaaa",
		Password: "aaaa",
		Tenant:   "tenant",
	}
	svc := NewAuditLogOverridesStep(repo,cfg)
	svc.fs = mm

	operation := internal.ProvisioningOperation{
		ProvisioningParameters: `{"ers_context": {"subaccount_id": "1234567890"}}`,
	}
	repo.InsertProvisioningOperation(operation)

	// when
	_, _, err := svc.Run(operation, NewLogDummy())
	//then
	require.Error(t, err)
	require.EqualError(t, err, "open /audit-log-script/script: file does not exist")

}

func TestAuditLog_HappyPath(t *testing.T) {
	// given
	mm := afero.NewMemMapFs()

	fileScript := `
func myScript() {
foo: sub_account_id
bar: tenant_id
return "fooBar"
}
`

	err := afero.WriteFile(mm, "/audit-log-script/script", []byte(fileScript), 0755)
	if err != nil {
		t.Fatalf("Unable to write contents to file: audit-log-script!!: %v", err)
	}

	repo := storage.NewMemoryStorage().Operations()
	cfg := auditlog.Config{
		URL:      "host1",
		User:     "aaaa",
		Password: "aaaa",
		Tenant:   "tenant",
	}
	svc := NewAuditLogOverridesStep(repo, cfg)
	svc.fs = mm

	inputCreatorMock := &automock.ProvisionInputCreator{}
	defer inputCreatorMock.AssertExpectations(t)
	expectedOverride := `
[FILTER]
        Name    lua
        Match   dex.*
        script  script.lua
        call    reformat

[FILTER]
        Name    lua
        Match   dex.*
        script  script.lua
        call    reformat
[OUTPUT]
        Name    stdout
        Match   dex.*
[OUTPUT]
        Name             http
        Match            dex.*
        Host             host1
        Port             8081
        URI              /audit-log/v2/security-events
        Header           content-type    application/json
        Header           Content-Type    text/plain
        HTTP_User        aaaa
        HTTP_Passwd      aaaa
        Format           json_stream
        tls              on
        tls.debug        1
`
	expectedFileScript := `
func myScript() {
foo: 1234567890
bar: tenant
return "fooBar"
}
`
	inputCreatorMock.On("AppendOverrides", "logging", []*gqlschema.ConfigEntryInput{
		{
			Key:   "fluent-bit.conf.script",
			Value: expectedFileScript,
		},
		{
			Key:   "fluent-bit.conf.extra",
			Value: expectedOverride,
		},
	}).Return(nil).Once()

	operation := internal.ProvisioningOperation{
		InputCreator:           inputCreatorMock,
		ProvisioningParameters: `{"ers_context": {"subaccount_id": "1234567890"}}`,
	}
	repo.InsertProvisioningOperation(operation)
	// when
	_, repeat, err := svc.Run(operation, NewLogDummy())
	//then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
}
