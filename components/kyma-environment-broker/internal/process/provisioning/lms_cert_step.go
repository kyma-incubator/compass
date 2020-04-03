package provisioning

import (
	"errors"
	"regexp"
	"time"

	"crypto/x509/pkix"

	"fmt"

	"encoding/base64"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	pollingInterval          = 15 * time.Second
	certPollingTimeout       = 4 * time.Minute
	tenantReadyRetryInterval = 30 * time.Second
	lmsTimeout               = 30 * time.Minute
)

type LmsClient interface {
	RequestCertificate(tenantID string, subject pkix.Name) (id string, privateKey []byte, err error)
	GetSignedCertificate(tenantID string, certID string) (cert string, found bool, err error)
	GetCACertificate(tenantID string) (cert string, found bool, err error)
	GetTenantStatus(tenantID string) (status lms.TenantStatus, err error)
	GetTenantInfo(tenantID string) (status lms.TenantInfo, err error)
}

type lmsCertStep struct {
	provider            LmsClient
	repo                storage.Operations
	normalizationRegexp *regexp.Regexp
}

func NewLmsCertificatesStep(certProvider LmsClient, os storage.Operations) *lmsCertStep {

	return &lmsCertStep{
		provider:            certProvider,
		repo:                os,
		normalizationRegexp: regexp.MustCompile("[^a-zA-Z0-9]+"),
	}
}

func (s *lmsCertStep) Name() string {
	return "Request_LMS_Certificates"
}

// Run executes getting LMS certificates steps, which means:
// 1. check if the tenant is ready
// 2. request certificates
// 3. poll CA and signed certificates
func (s *lmsCertStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if operation.Lms.Failed {
		logger.Info("LMS has failed, skipping")
		return operation, 0, nil
	}

	if operation.Lms.TenantID == "" {
		logger.Error("Create LMS Tenant step must be run before")
		return operation, 0, errors.New("the step needs to be run after 'Create LMS tenant' step")
	}

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		logger.Errorf("Unable to get provisioning parameters", err.Error())
		return operation, 0, errors.New("unable to get provisioning parameters")
	}

	// check if LMS tenant is ready
	status, err := s.provider.GetTenantStatus(operation.Lms.TenantID)
	if err != nil {
		logger.Errorf("Unable to get LMS Tenant status: %s", err.Error())
		if time.Since(operation.Lms.RequestedAt) > lmsTimeout {
			logger.Error("Setting LMS operation failed - tenant provisioning timed out, last error: %s", err.Error())
			return s.failLmsAndUpdate(operation)
		}
		return operation, tenantReadyRetryInterval, nil
	}
	if !(status.ElasticsearchDNSResolves && status.KibanaDNSResolves) {
		logger.Infof("LMS tenant not ready: elasticDNS=%v, kibanaDNS=%v", status.ElasticsearchDNSResolves, status.KibanaDNSResolves)
		if time.Since(operation.Lms.RequestedAt) > lmsTimeout {
			logger.Error("Setting LMS operation failed - tenant provisioning timed out")
			return s.failLmsAndUpdate(operation)
		}
		return operation, tenantReadyRetryInterval, nil
	}

	tenantInfo, err := s.provider.GetTenantInfo(operation.Lms.TenantID)
	if err != nil {
		logger.Errorf("Unable to get LMS Tenant info: %s", err.Error())
		if time.Since(operation.Lms.RequestedAt) > lmsTimeout {
			logger.Error("Setting LMS operation failed - tenant provisioning timed out, last error: %s", err.Error())
			return s.failLmsAndUpdate(operation)
		}
		return operation, tenantReadyRetryInterval, nil
	}

	// request certificates
	subj := pkix.Name{
		CommonName:         "fluentbit", // do not modify
		Organization:       []string{pp.ErsContext.GlobalAccountID},
		OrganizationalUnit: []string{pp.ErsContext.SubAccountID},
	}
	certId, pKey, err := s.provider.RequestCertificate(operation.Lms.TenantID, subj)
	if err != nil {
		logger.Errorf("Unable to request LMS Certificates %s", err.Error())
		return operation, 5 * time.Second, nil
	}

	var signedCert string
	var caCert string

	// certs cannot be stored so there is a need to poll until certs are ready
	// get Signed Certificate
	err = wait.PollImmediate(pollingInterval, certPollingTimeout, func() (done bool, err error) {
		c, found, err := s.provider.GetSignedCertificate(operation.Lms.TenantID, certId)
		if err != nil {
			logger.Warnf("Unable to get LMS Signed Certificate: %s, retrying", err.Error())
			return false, nil
		}
		if !found {
			logger.Info("LMS Signed Certificate not ready")
			return false, nil
		}
		signedCert = c
		return true, nil
	})
	if err != nil {
		logger.Errorf("Setting LMS operation failed: %s", err.Error())
		return s.failLmsAndUpdate(operation)
	}

	// get CA cert
	err = wait.PollImmediate(pollingInterval, certPollingTimeout, func() (done bool, err error) {
		c, found, err := s.provider.GetCACertificate(operation.Lms.TenantID)
		if err != nil {
			logger.Warnf("Unable to get LMS CA Certificate: %s", err.Error())
			return false, nil
		}
		if !found {
			logger.Info("LMS Ca Certificate not ready")
			return false, nil
		}
		caCert = c
		return true, nil
	})
	if err != nil {
		logger.Errorf("Setting LMS operation failed: %s", err.Error())
		return s.failLmsAndUpdate(operation)
	}

	operation.InputCreator.AppendOverrides("logging", []*gqlschema.ConfigEntryInput{
		{Key: "fluent-bit.conf.Output.forward.enabled", Value: "true"},
		{Key: "fluent-bit.backend.forward.host", Value: fmt.Sprintf("forward.%s", tenantInfo.DNS)},
		{Key: "fluent-bit.backend.forward.port", Value: "8443"},
		{Key: "fluent-bit.backend.forward.tls.enabled", Value: "true"},
		{Key: "fluent-bit.backend.forward.tls.verify", Value: "On"},

		// certs and private key must be encoded by base64
		{Key: "fluent-bit.backend.forward.tls.ca", Value: base64.StdEncoding.EncodeToString([]byte(caCert))},
		{Key: "fluent-bit.backend.forward.tls.cert", Value: base64.StdEncoding.EncodeToString([]byte(signedCert))},
		{Key: "fluent-bit.backend.forward.tls.key", Value: base64.StdEncoding.EncodeToString(pKey)},

		{Key: "fluent-bit.conf.extra", Value: fmt.Sprintf(`
[FILTER]
        Name record_modifier
        Match *
        Record cluster_name %s
`, pp.ErsContext.SubAccountID)}, // cluster_name is a tag added to log entry, allows to filter logs by a cluster
	})

	return operation, 0, nil
}

func (s *lmsCertStep) failLmsAndUpdate(operation internal.ProvisioningOperation) (internal.ProvisioningOperation, time.Duration, error) {
	operation.Lms.Failed = true
	modifiedOp, err := s.repo.UpdateProvisioningOperation(operation)
	if err != nil {
		// update has failed - retry after 0.5 sec
		return operation, 500 * time.Millisecond, nil
	}
	return *modifiedOp, 0, nil
}
