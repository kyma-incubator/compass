package provisioning

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type AuditLogOverrides struct {
	operationManager *process.ProvisionOperationManager
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

	return &AuditLogOverrides{
		process.NewProvisionOperationManager(os),
	}
}

func (alo *AuditLogOverrides) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	// fetch the username, url and password
	//file, err := os.Open("audit-log-config")

	alcFile, err := readFile("audit-log-config")
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

	//luaScript, err := ioutil.ReadFile("audit-config-script")
	luaScript, err := readFile("audit-log-script")
	if err != nil {
		logger.Errorf("Unable to read audit config script: %v", err)
		return operation, 0, nil
	}

	// Fetch the region
	region := "east"
	var c aduditLogCred
	for _, a := range alc {
		if v, ok := a[region]; ok {
			c = v
			break
		} else {
			logger.Errorf("Unable to find credentials for the audit log for the region: %v", region)
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

func readFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}
