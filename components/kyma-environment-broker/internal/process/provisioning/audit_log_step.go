package provisioning

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/auditlog"

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
	luaScript, err := alo.readFile("/auditlog-script/script")
	if err != nil {
		logger.Errorf("Unable to read audit config script: %v", err)
		return operation, 0, err
	}

	replaceSubAccountID := strings.Replace(string(luaScript), "sub_account_id", pp.ErsContext.SubAccountID, -1)
	replaceTenantID := strings.Replace(replaceSubAccountID, "tenant_id", alo.auditLogConfig.Tenant, -1)

	u, err := url.Parse(alo.auditLogConfig.URL)
	if err != nil {
		logger.Errorf("Unable to parse the URL: %v", err.Error())
		return operation, 0, err
	}
	if u.Path == "" {
		logger.Errorf("There is no Path passed in the URL")
		return operation, 0, errors.New("there is no Path passed in the URL")
	}
	auditLogHost, auditLogPort, err := net.SplitHostPort(u.Host)
	if err != nil {
		logger.Errorf("Unable to split URL: %v", err.Error())
		return operation, 0, err
	}
	if auditLogPort == "" {
		auditLogPort = "443"
		logger.Infof("There is no Port passed in the URL. Setting default to 443")
	}

	operation.InputCreator.AppendOverrides("logging", []*gqlschema.ConfigEntryInput{
		{Key: "fluent-bit.conf.script", Value: replaceTenantID},
		{Key: "fluent-bit.conf.extra", Value: fmt.Sprintf(`
[INPUT]
        Name              tail
        Tag               dex.*
        Path              /var/log/containers/*_dex-*.log
        DB                /var/log/flb_kube_dex.db
        parser            docker
        Mem_Buf_Limit     5MB
        Skip_Long_Lines   On
        Refresh_Interval  10
[FILTER]
        Name    lua
        Match   dex.*
        script  script.lua
        call    reformat
[OUTPUT]
        Name             http
        Match            dex.*
        Retry_Limit      False
        Host             %s
        Port             %s
        URI              %ssecurity-events
        Header           Content-Type application/json
        HTTP_User        %s
        HTTP_Passwd      %s
        Format           json_stream
        tls              on
`, auditLogHost, auditLogPort, u.Path, alo.auditLogConfig.User, alo.auditLogConfig.Password)},
		{Key: "fluent-bit.externalServiceEntry.resolution", Value: "DNS"},
		{Key: "fluent-bit.externalServiceEntry.hosts", Value: fmt.Sprintf(`- %s`, auditLogHost)},
		{Key: "fluent-bit.externalServiceEntry.ports", Value: fmt.Sprintf(`- number: %s
  name: https
  protocol: TLS`, auditLogPort)},
	})
	return operation, 0, nil
}

func (alo *AuditLogOverrides) readFile(fileName string) ([]byte, error) {
	return afero.ReadFile(alo.fs, fileName)
}
