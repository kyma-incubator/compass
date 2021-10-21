package cert

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"go.mozilla.org/pkcs7"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// CertSvcConfig is the configuration needed for getting a certificate from cert service
type CertSvcConfig struct {
	SubjectPattern string
	CommonName     string
	Locality       string
	Policy         string
	CSREndpoint    string

	ClientID     string
	ClientSecret string
	OAuthURL     string

	TokenPath      string        `envconfig:"default=/oauth/token"`
	CertSvcAPIPath string        `envconfig:"default=/v3/synchronous/certificate"`
	IssuerLocality string        `envconfig:"default=local"`
	RetryAttempts  uint          `envconfig:"default=2"`
	RetryDelay     time.Duration `envconfig:"default=100ms"`
}

// Validate validates a cert service config
func (csc *CertSvcConfig) Validate() error {
	if len(csc.SubjectPattern) == 0 {
		return errors.New("Subject Pattern cannot be empty")
	}
	if len(csc.CommonName) == 0 {
		return errors.New("CommonName cannot be empty")
	}
	if len(csc.Locality) == 0 {
		return errors.New("Locality cannot be empty")
	}
	if len(csc.Policy) == 0 {
		return errors.New("Policy cannot be empty")
	}
	if len(csc.CSREndpoint) == 0 {
		return errors.New("CSREndpoint cannot be empty")
	}
	if len(csc.ClientID) == 0 {
		return errors.New("ClientID cannot be empty")
	}
	if len(csc.ClientSecret) == 0 {
		return errors.New("ClientSecret cannot be empty")
	}
	if len(csc.OAuthURL) == 0 {
		return errors.New("OAuthURL cannot be empty")
	}

	return nil
}

type csrRequest struct {
	Csr    csrPayload `json:"csr"`
	Policy string     `json:"policy"`
}

type subjectElement struct {
	Name  string `json:"shortName"`
	Value string `json:"value"`
}

type subject struct {
	Subject []subjectElement `json:"subject"`
}

type csrPayload struct {
	Value              string  `json:"value"`
	Type               string  `json:"type"`
	SubjectReplacement subject `json:"replace"`
}

type csrResponse struct {
	CrtResponse crtResponse `json:"certificateChain"`
}

type crtResponse struct {
	Crt string `json:"value"`
}

type client struct {
	config CertSvcConfig

	*http.Client
}

// NewCertSvcClient returns a certificate service client
func NewCertSvcClient(base *http.Client, config CertSvcConfig) *client {
	return &client{
		Client: base,
		config: config,
	}
}

// IssueClientCert issues a client certificate from the cert service.
func (c *client) IssueClientCert(ctx context.Context) (*tls.Certificate, error) {
	if err := c.config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            c.stringSliceOrNil(GetCountry(c.config.SubjectPattern)),
			Organization:       c.stringSliceOrNil(GetOrganization(c.config.SubjectPattern)),
			OrganizationalUnit: GetAllOrganizationalUnits(c.config.SubjectPattern),
			Locality:           []string{c.config.Locality},
			Province:           c.stringSliceOrNil(GetProvince(c.config.SubjectPattern)),
			CommonName:         c.config.CommonName,
		},
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return nil, err
	}

	pemEncodedCSR := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr,
	})

	csrRequest := csrRequest{
		Csr: csrPayload{
			Value: string(pemEncodedCSR),
			Type:  "pkcs10-pem",
			SubjectReplacement: subject{
				c.subjectPatternToSubjectElementSlice(),
			},
		},
		Policy: c.config.Policy,
	}

	body, err := json.Marshal(csrRequest)
	if err != nil {
		return nil, err
	}

	ctxWithClient := context.WithValue(ctx, oauth2.HTTPClient, c.Client)
	ccConf := clientcredentials.Config{
		ClientID:     c.config.ClientID,
		ClientSecret: c.config.ClientSecret,
		TokenURL:     c.config.OAuthURL + c.config.TokenPath,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}
	client := ccConf.Client(ctxWithClient)

	cert := &tls.Certificate{}
	err = retry.Do(func() error {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.config.CSREndpoint+c.config.CertSvcAPIPath, bytes.NewBuffer(body))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.C(ctx).Info("Failed to close HTTP response body")
			}
		}()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			log.C(ctx).Errorf("Unexpected status code while issuing client cert: Status Code: %d Body: %s", resp.StatusCode, bodyBytes)
			return errors.Errorf("unexpected status code while issuing client cert %d", resp.StatusCode)
		}

		csrResp := &csrResponse{}
		err = json.Unmarshal(bodyBytes, csrResp)
		if err != nil {
			return err
		}

		pemCert, err := c.decodePem(csrResp.CrtResponse.Crt)
		if err != nil {
			return err
		}

		cert = &tls.Certificate{
			Certificate: pemCert,
			PrivateKey:  privateKey,
		}

		// By doc, the first element is always the "leaf"/client cert
		parsedClientCert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return err
		}

		if len(parsedClientCert.Issuer.Locality) > 0 {
			issuerLocality := parsedClientCert.Issuer.Locality[0]
			if issuerLocality != c.config.IssuerLocality {
				log.C(ctx).Errorf("Issuer locality of the client cert: %s is not the desired one: %s. Will try issuing a certificate again...", issuerLocality, c.config.IssuerLocality)
				return errors.Errorf("issuer locality of the client cert does not match the expected one: %s", c.config.IssuerLocality)
			}
		}
		return nil
	}, retry.Attempts(c.config.RetryAttempts), retry.Delay(c.config.RetryDelay))
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *client) decodePem(chain string) ([][]byte, error) {
	res := make([][]byte, 0)

	pemBlock, _ := pem.Decode([]byte(chain))

	pkcs, err := pkcs7.Parse(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	for _, c := range pkcs.Certificates {
		res = append(res, c.Raw)
	}

	return res, nil
}

func (c *client) subjectPatternToSubjectElementSlice() []subjectElement {
	result := []subjectElement{
		{
			Name:  "C",
			Value: GetCountry(c.config.SubjectPattern),
		},
		{
			Name:  "O",
			Value: GetOrganization(c.config.SubjectPattern),
		},
	}

	result = append(result, c.stringSliceToSubjectElementSlice("OU", GetAllOrganizationalUnits(c.config.SubjectPattern))...)
	result = append(result, subjectElement{
		Name:  "L",
		Value: c.config.Locality,
	}, subjectElement{
		Name:  "CN",
		Value: c.config.CommonName,
	})

	return result
}

func (c *client) stringSliceToSubjectElementSlice(name string, values []string) []subjectElement {
	result := make([]subjectElement, 0, len(values))
	for _, val := range values {
		result = append(result, subjectElement{
			Name:  name,
			Value: val,
		})
	}
	return result
}

func (c *client) stringSliceOrNil(str string) []string {
	if len(str) == 0 {
		return nil
	}
	return []string{str}
}
