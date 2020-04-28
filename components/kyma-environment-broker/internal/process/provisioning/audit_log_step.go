package provisioning

import (
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type AuditLogOverrides struct {
	operationManager *process.ProvisionOperationManager
	fs               afero.Fs
}

func (alo *AuditLogOverrides) Name() string {
	return "Audit_Log_Overrides"
}

type aduditLogCred struct {
	Host     string `yaml:"host"`
	HTTPUser string `yaml:"http-user"`
	HTTPPwd  string `yaml:"http-pwd"`
}

func NewAuditLogOverridesStep(os storage.Operations) *AuditLogOverrides {
	fileSystem := afero.NewOsFs()

	return &AuditLogOverrides{
		process.NewProvisionOperationManager(os),
		fileSystem,
	}
}

func (alo *AuditLogOverrides) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	// fetch the username, url and password
	//file, err := os.Open("audit-log-config")

	alcFile, err := alo.readFile("audit-log-config")
	if err != nil {
		logger.Errorf("Unable to read audit log config file: %v", err)
		return operation, 0, err
	}
	var alc []map[string]aduditLogCred
	err = yaml.Unmarshal(alcFile, &alc)
	if err != nil {
		logger.Errorf("Error parsing audit log config file: %v", err)
		return operation, 0, err
	}

	luaScript, err := alo.readFile("audit-log-script")
	if err != nil {
		logger.Errorf("Unable to read audit config script: %v", err)
		return operation, 0, err
	}
	// Fetch the region
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		logger.Errorf("Unable to get provisioning parameters", err.Error())
		return operation, 0, errors.New("unable to get provisioning parameters")
	}

	var c aduditLogCred
	for _, a := range alc {
		if v, ok := a[pp.PlatformRegion]; ok {
			c = v
			break
		} else {
			logger.Errorf("Unable to find credentials for the audit log for the region: %v", pp.PlatformRegion)
			return operation, 0, nil
		}
	}

	operation.InputCreator.AppendOverrides("logging", []*gqlschema.ConfigEntryInput{
		{Key: "fluent-bit.conf.script", Value: string(luaScript)},
		{Key: "fluent-bit.conf.extra", Value: fmt.Sprintf(`
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
        Host             %s
        Port             8081
        URI              /audit-log/v2/security-events
        Header           content-type    application/json 
        Header           Content-Type    text/plain
        HTTP_User        %s
        HTTP_Passwd      %s
        Format           json_stream
        tls              on
        tls.debug        1

`, c.Host, c.HTTPUser, c.HTTPPwd)},
	})
	return operation, 0, nil
}

func (alo *AuditLogOverrides) readFile(fileName string) ([]byte, error) {
	return afero.ReadFile(alo.fs, fileName)
}
