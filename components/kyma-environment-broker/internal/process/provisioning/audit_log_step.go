package provisioning

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/auditlog"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type AuditLogOverrides struct {
	operationManager *process.ProvisionOperationManager
	fs               afero.Fs
	auditLogConfig   auditlog.Config
}

func (alo *AuditLogOverrides) Name() string {
	return "Audit_Log_Overrides"
}

func NewAuditLogOverridesStep(os storage.Operations, cfg auditlog.Config) *AuditLogOverrides {
	fileSystem := afero.NewOsFs()

	return &AuditLogOverrides{
		process.NewProvisionOperationManager(os),
		fileSystem,
		cfg,
	}
}

func (alo *AuditLogOverrides) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	// Fetch the region
	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		logger.Errorf("Unable to get provisioning parameters", err.Error())
		return operation, 0, errors.New("unable to get provisioning parameters")
	}
	luaScript, err := alo.readFile("/audit-log-script/script")
	if err != nil {
		logger.Errorf("Unable to read audit config script: %v", err)
		return operation, 0, err
	}

	replaceSubAccountID := strings.Replace(string(luaScript), "sub_account_id", pp.ErsContext.SubAccountID, -1)
	replaceTenantID := strings.Replace(replaceSubAccountID, "tenant_id", alo.auditLogConfig.Tenant, -1)

	operation.InputCreator.AppendOverrides("logging", []*gqlschema.ConfigEntryInput{
		{Key: "fluent-bit.conf.script", Value: replaceTenantID},
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
        call    reformat
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
`, alo.auditLogConfig.URL, alo.auditLogConfig.User, alo.auditLogConfig.Password)},
		{Key: "fluent-bit.externalServiceEntry.resolution", Value: "DNS"},
		{Key: "fluent-bit.externalServiceEntry.hosts", Value: fmt.Sprintf("%s",alo.auditLogConfig.URL)},
		{Key: "fluent-bit.externalServiceEntry.ports", Value: `
         - number: 8081
           name: https
           protocol: TLS`},

	})
	return operation, 0, nil
}

func (alo *AuditLogOverrides) readFile(fileName string) ([]byte, error) {
	return afero.ReadFile(alo.fs, fileName)
}
