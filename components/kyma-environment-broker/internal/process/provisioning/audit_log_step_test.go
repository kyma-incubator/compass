package provisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestAuditLog_ConfigFileDoesNotExist(t *testing.T) {
	// given

	memoryStorage := storage.NewMemoryStorage()
	svc := NewAuditLogOverridesStep(memoryStorage.Operations())
	svc.fs = afero.NewMemMapFs()

	operation := internal.ProvisioningOperation{}

	// when
	_, _, err := svc.Run(operation, NewLogDummy())
	//then
	require.Error(t, err)
	require.EqualError(t, err, "open audit-log-config: file does not exist")

}

func TestAuditLog_ScriptFileDoesNotExist(t *testing.T) {
	// given
	mm := afero.NewMemMapFs()
	_, err := mm.Create("audit-log-config")
	if err != nil {
		t.Fatalf("Unable to create file: audit-log-config!!: %v", err)
	}
	fileData := `[
   {
      "east": {
         "host": "host",
         "http-user": "aaaa",
         "http-pwd": "aaaa"
      }
   },
   {
      "west": {
         "host": "host",
         "http-user": "bbbb",
         "http-pwd": "bbbb"
      }
   }
]`

	fYaml, err := yaml.JSONToYAML([]byte(fileData))
	if err != nil {
		t.Fatalf("Unable to convert to yaml: %v", err)
	}
	err = afero.WriteFile(mm, "audit-log-config", fYaml, 0755)
	if err != nil {
		t.Fatalf("Unable to write contents to file: audit-log-config!!: %v", err)
	}

	memoryStorage := storage.NewMemoryStorage()
	svc := NewAuditLogOverridesStep(memoryStorage.Operations())
	svc.fs = mm

	operation := internal.ProvisioningOperation{}

	// when
	_, _, err = svc.Run(operation, NewLogDummy())
	//then
	require.Error(t, err)
	require.EqualError(t, err, "open audit-log-script: file does not exist")

}

func TestAuditLog_HappyPath(t *testing.T) {
	// given
	mm := afero.NewMemMapFs()
	_, err := mm.Create("audit-log-config")
	if err != nil {
		t.Fatalf("Unable to create file: audit-log-config!!: %v", err)
	}
	fileData := `[
   {
      "east": {
         "host": "host1",
         "http-user": "aaaa",
         "http-pwd": "aaaa"
      }
   },
   {
      "west": {
         "host": "host2",
         "http-user": "bbbb",
         "http-pwd": "bbbb"
      }
   }
]`

	fileScript := `
func myScript() {
foo: sub_account_id
return "fooBar"
}
`
	fyaml, err := yaml.JSONToYAML([]byte(fileData))
	if err != nil {
		t.Fatalf("Unable to convert to yaml: %v", err)
	}
	err = afero.WriteFile(mm, "audit-log-config", fyaml, 0755)
	if err != nil {
		t.Fatalf("Unable to write contents to file: audit-log-config!!: %v", err)
	}

	err = afero.WriteFile(mm, "audit-log-script", []byte(fileScript), 0755)
	if err != nil {
		t.Fatalf("Unable to write contents to file: audit-log-script!!: %v", err)
	}

	repo := storage.NewMemoryStorage().Operations()
	svc := NewAuditLogOverridesStep(repo)
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
        call    append_uuid
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
		ProvisioningParameters: `{"platform_region": "east", "ers_context": {"subaccount_id": "1234567890"}}`,
	}
	repo.InsertProvisioningOperation(operation)
	// when
	_, repeat, err := svc.Run(operation, NewLogDummy())
	//then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
}
