package lms

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/iosafety"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Client interface {
	CreateTenant(input CreateTenantInput) (o CreateTenantOutput, err error)
	GetTenantStatus(tenantID string) (status TenantStatus, err error)
	GetTenantInfo(tenantID string) (status TenantInfo, err error)

	GetCACertificate(tenantID string) (cert string, found bool, err error)
	GetCertificateByURL(url string) (cert string, found bool, err error)
	RequestCertificate(tenantID string, subject pkix.Name) (string, []byte, error)
}

// ClusterType can be ha or single-node
type ClusterType string

type client struct {
	url string

	// tenant predefined values
	clusterType ClusterType
	environment string

	token      string
	samlTenant string

	log logrus.FieldLogger
}

const (
	ClusterTypeHA         ClusterType = "ha"
	ClusterTypeSingleNode ClusterType = "single-node"
)

type Config struct {
	URL string

	// tenant predefined values
	ClusterType ClusterType // ha or single-node
	Environment string

	Token      string
	SamlTenant string
	Region     string `envconfig:"optional"`
	Mandatory  bool   `envconfig:"default=true"`

	Disabled bool
}

func (c Config) Validate() error {
	if c.ClusterType != ClusterTypeSingleNode && c.ClusterType != ClusterTypeHA {
		return fmt.Errorf("unknown cluster type '%s'", c.ClusterType)
	}
	return nil
}

func NewClient(cfg Config, log logrus.FieldLogger) Client {
	return &client{
		url:         cfg.URL,
		clusterType: cfg.ClusterType,
		environment: cfg.Environment,
		token:       cfg.Token,
		samlTenant:  cfg.SamlTenant,
		log:         log,
	}
}

type createTenantPayload struct {
	Name            string   `json:"name"`
	Region          string   `json:"region"`
	SamlGroups      []string `json:"samlGroups"`
	ClusterType     string   `json:"clusterType"`
	DataCenter      string   `json:"datacenter"`
	Environment     string   `json:"environment"`
	Project         string   `json:"project"`
	SamlTenant      string   `json:"samlTenant"`
	Costcenter      int      `json:"costcenter"`
	GlobalAccountID string   `json:"globalAccountID"`
}

type CreateTenantInput struct {
	Name            string
	Region          string
	GlobalAccountID string
}

type CreateTenantOutput struct {
	ID string `json:"id"`
}

type TenantStatus struct {
	KibanaDNSResolves        bool   `json:"kibanaDNSResolves"`
	ElasticsearchDNSResolves bool   `json:"elasticsearchDNSResolves"`
	KibanaState              string `json:"kibanaState"`
}

type TenantInfo struct {
	ID  string `json:"id"`
	DNS string `json:"dns"`
}

// CreateTenant create the LMS tenant
// Tenant creation means creation of a cluster, which must be reusable for the same tenant/region/project
func (c *client) CreateTenant(input CreateTenantInput) (o CreateTenantOutput, err error) {
	payload := createTenantPayload{
		Name:        input.Name,
		Region:      input.Region,
		SamlGroups:  []string{"skr-logging-viewer"},
		ClusterType: string(c.clusterType),
		DataCenter:  "all",
		Environment: string(c.environment),
		Project:     "kyma", // the tenant will use always the same project
		SamlTenant:  c.samlTenant,
		Costcenter:  0,
	}
	jsonPayload, err := json.Marshal(payload)
	c.log.Debugf("Create tenant payload: %s", string(jsonPayload))
	if err != nil {
		return CreateTenantOutput{}, errors.Wrapf(err, "while encoding Create Tenant payload")
	}

	url := fmt.Sprintf("%s/tenants", c.url)
	c.log.Debugf("url: %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return CreateTenantOutput{}, errors.Wrapf(err, "while creating request Create Tenant")
	}
	req.Header.Add("X-LMS-Token", c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CreateTenantOutput{}, errors.Wrapf(err, "while calling Create Tenant endpoint")
	}
	defer func() {
		if drainErr := iosafety.DrainReader(resp.Body); drainErr != nil {
			err = multierror.Append(err, errors.Wrap(drainErr, "while trying to drain body reader"))
		}

		if closeErr := resp.Body.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "while trying to close body reader"))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	var output struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	err = json.Unmarshal(body, &output)
	if err != nil {
		return CreateTenantOutput{}, errors.Wrapf(err, "while unmarshalling response: %s", string(body))
	}

	if resp.StatusCode >= 400 {
		return CreateTenantOutput{}, errors.Errorf("error when calling create tenant endpoint,"+
			" status code: %d, error: '%s' message: '%s'",
			resp.StatusCode, output.Error, output.Message)
	}

	return CreateTenantOutput{ID: output.ID}, nil
}

func (c *client) GetTenantStatus(tenantID string) (status TenantStatus, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tenants/%s/status", c.url, tenantID), nil)
	if err != nil {
		return TenantStatus{}, errors.Wrapf(err, "while creating Get Tenant Status request")
	}
	req.Header.Add("X-LMS-Token", c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TenantStatus{}, errors.Wrapf(err, "while calling Get Tenant Status endpoint")
	}
	defer func() {
		if drainErr := iosafety.DrainReader(resp.Body); drainErr != nil {
			err = multierror.Append(err, errors.Wrap(drainErr, "while trying to drain body reader"))
		}

		if closeErr := resp.Body.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "while trying to close body reader"))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	var tenantStatus TenantStatus

	err = json.Unmarshal(body, &tenantStatus)
	if err != nil {
		return TenantStatus{}, errors.Wrapf(err, "while unmarshalling response: %s", string(body))
	}

	if resp.StatusCode >= 400 {
		return TenantStatus{}, errors.Errorf("error when calling get tenant status endpoint,"+
			" status code: %d, body: %s",
			resp.StatusCode, body)
	}

	return tenantStatus, nil
}

func (c *client) GetTenantInfo(tenantID string) (status TenantInfo, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tenants/%s", c.url, tenantID), nil)
	if err != nil {
		return TenantInfo{}, errors.Wrapf(err, "while creating Get Tenant request")
	}
	req.Header.Add("X-LMS-Token", c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TenantInfo{}, errors.Wrapf(err, "while calling Get Tenant endpoint")
	}
	defer func() {
		if drainErr := iosafety.DrainReader(resp.Body); drainErr != nil {
			err = multierror.Append(err, errors.Wrap(drainErr, "while trying to drain body reader"))
		}

		if closeErr := resp.Body.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "while trying to close body reader"))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	var response TenantInfo

	err = json.Unmarshal(body, &response)
	if err != nil {
		return TenantInfo{}, errors.Wrapf(err, "while unmarshalling response: %s", string(body))
	}

	if resp.StatusCode >= 400 {
		return TenantInfo{}, errors.Errorf("error when calling get tenant info endpoint,"+
			" status code: %d, body: %s",
			resp.StatusCode, body)
	}

	return response, nil
}

func (c *client) GetCertificateByURL(url string) (cert string, found bool, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", false, errors.Wrapf(err, "while creating Get Certificate request (%s)", url)
	}
	req.Header.Add("X-LMS-Token", c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, errors.Wrapf(err, "while calling Get Certificate endpoint (%s)", url)
	}
	defer func() {
		if drainErr := iosafety.DrainReader(resp.Body); drainErr != nil {
			err = multierror.Append(err, errors.Wrap(drainErr, "while trying to drain body reader"))
		}

		if closeErr := resp.Body.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "while trying to close body reader"))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	var certResponse struct {
		Cert string `json:"cert"`
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}

	if resp.StatusCode >= 400 {
		return "", false, errors.Errorf("error when calling get tenant status endpoint,"+
			" status code: %d, body: %s",
			resp.StatusCode, body)
	}

	err = json.Unmarshal(body, &certResponse)
	if err != nil {
		return "", false, errors.Wrapf(err, "while unmarshalling response: %s", string(body))
	}

	return certResponse.Cert, true, nil
}

func (c *client) getCertificate(tenantID string, certID string) (cert string, found bool, err error) {
	return c.GetCertificateByURL(fmt.Sprintf("%s/tenants/%s/certs/%s", c.url, tenantID, certID))
}

func (c *client) GetCACertificate(tenantID string) (cert string, found bool, err error) {
	return c.GetCertificateByURL(fmt.Sprintf("%s/tenants/%s/certs/ca", c.url, tenantID))
}

func (c *client) RequestCertificate(tenantID string, subject pkix.Name) (string, []byte, error) {
	csr, privateKey, err := c.generateCSR(subject)
	if err != nil {
		return "", privateKey, errors.Wrapf(err, "while generating CSR")
	}
	var payload struct {
		CertId int    `json:"certId"`
		Csr    string `json:"csr"`
	}
	payload.Csr = string(csr)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", privateKey, errors.Wrapf(err, "while encoding Create Request payload")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/tenants/%s/certs", c.url, tenantID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", privateKey, err
	}
	req.Header.Add("X-LMS-Token", c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", privateKey, err
	}
	defer func() {
		if drainErr := iosafety.DrainReader(resp.Body); drainErr != nil {
			err = multierror.Append(err, errors.Wrap(drainErr, "while trying to drain body reader"))
		}

		if closeErr := resp.Body.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "while trying to close body reader"))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	var response struct {
		CallbackURL string `json:"callbackUrl"`
	}

	if resp.StatusCode >= 400 {
		return "", []byte{}, errors.Errorf("error when calling request cert endpoint,"+
			" status code: %d, body: %s",
			resp.StatusCode, body)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", []byte{}, errors.Wrapf(err, "while unmarshalling response: %s", string(body))
	}
	return response.CallbackURL, privateKey, nil
}

func (c *client) generateCSR(subject pkix.Name) (csr []byte, privateKey []byte, err error) {
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 2048)

	template := x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	csrBuf := bytes.Buffer{}
	pkBuf := bytes.Buffer{}
	if err := pem.Encode(&csrBuf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		return nil, nil, err
	}
	if err := pem.Encode(&pkBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keyBytes)}); err != nil {
		return nil, nil, err
	}

	return csrBuf.Bytes(), pkBuf.Bytes(), nil
}
