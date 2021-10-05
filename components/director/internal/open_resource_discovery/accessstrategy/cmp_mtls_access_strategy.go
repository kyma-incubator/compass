package accessstrategy

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
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.mozilla.org/pkcs7"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"sync"
)

const cmpMTLSConfigPrefix = "ACCESS_STRATEGY_SAP_CMP_MTLS_V1"

type cmpMTLSAccessStrategyExecutor struct {
	lock sync.RWMutex

	client *http.Client

	config *cmpMTLSConfig
}

func newCMPmTLSAccessStrategyExecutor() *cmpMTLSAccessStrategyExecutor {
	return &cmpMTLSAccessStrategyExecutor{
		lock: sync.RWMutex{},
	}
}

type cmpMTLSConfig struct {
	SubjectPattern string
	CommonName     string
	Locality       string
	Policy         string
	CSREndpoint    string

	ClientID     string
	ClientSecret string
	OAuthURL     string
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

// Execute performs the access strategy's specific execution logic
func (as *cmpMTLSAccessStrategyExecutor) Execute(ctx context.Context, baseClient *http.Client, documentURL string) (*http.Response, error) {
	if !as.isInitialized() {
		if err := as.initialize(ctx, baseClient); err != nil {
			return nil, errors.Wrap(err, "while initializing access strategy sap:cmp-mtls:v1")
		}
	}
	return as.client.Get(documentURL)
}

func (as *cmpMTLSAccessStrategyExecutor) isInitialized() bool {
	as.lock.RLock()
	defer as.lock.RUnlock()

	return as.config != nil && as.client != nil
}

func (as *cmpMTLSAccessStrategyExecutor) initialize(ctx context.Context, baseClient *http.Client) error {
	as.lock.Lock()
	defer as.lock.Unlock()

	if as.config == nil {
		cfg := cmpMTLSConfig{}
		if err := envconfig.InitWithPrefix(&cfg, cmpMTLSConfigPrefix); err != nil {
			return err
		}
		as.config = &cfg
	}

	if as.client == nil {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		csrTemplate := x509.CertificateRequest{
			Subject: pkix.Name{
				Country:            as.stringSliceOrNil(cert.GetCountry(as.config.SubjectPattern)),
				Organization:       as.stringSliceOrNil(cert.GetOrganization(as.config.SubjectPattern)),
				OrganizationalUnit: cert.GetAllOrganizationalUnits(as.config.SubjectPattern),
				Locality:           []string{as.config.Locality},
				Province:           as.stringSliceOrNil(cert.GetProvince(as.config.SubjectPattern)),
				CommonName:         as.config.CommonName,
			},
		}

		csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
		if err != nil {
			return err
		}

		pemEncodedCSR := pem.EncodeToMemory(&pem.Block{
			Type: "CERTIFICATE REQUEST", Bytes: csr,
		})

		csrRequest := csrRequest{
			Csr: csrPayload{
				Value: string(pemEncodedCSR),
				Type:  "pkcs10-pem",
				SubjectReplacement: subject{
					as.subjectPatternToSubjectElementSlice(),
				},
			},
			Policy: as.config.Policy,
		}

		body, err := json.Marshal(csrRequest)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, as.config.CSREndpoint, bytes.NewBuffer(body))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		ctx := context.WithValue(ctx, oauth2.HTTPClient, baseClient)
		ccConf := clientcredentials.Config{
			ClientID:     as.config.ClientID,
			ClientSecret: as.config.ClientSecret,
			TokenURL:     as.config.OAuthURL,
			AuthStyle:    oauth2.AuthStyleAutoDetect,
		}

		client := ccConf.Client(ctx)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

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

		clientCert := tls.Certificate{
			Certificate: as.decodePem(csrResp.CrtResponse.Crt),
			PrivateKey:  privateKey,
		}

		tr := baseClient.Transport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}

		as.client = &http.Client{
			Timeout:   baseClient.Timeout,
			Transport: tr,
		}
	}

	return nil
}

func (as *cmpMTLSAccessStrategyExecutor) decodePem(chain string) [][]byte {
	res := make([][]byte, 0, 0)

	pemBlock, _ := pem.Decode([]byte(chain))
	pkcs, _ := pkcs7.Parse(pemBlock.Bytes)
	for _, c := range pkcs.Certificates {
		res = append(res, c.Raw)
	}

	return res
}

func (as *cmpMTLSAccessStrategyExecutor) subjectPatternToSubjectElementSlice() []subjectElement {
	result := []subjectElement{
		{
			Name:  "C",
			Value: cert.GetCountry(as.config.SubjectPattern),
		},
		{
			Name:  "O",
			Value: cert.GetOrganization(as.config.SubjectPattern),
		},
	}

	result = append(result, as.stringSliceToSubjectElementSlice("OU", cert.GetAllOrganizationalUnits(as.config.SubjectPattern))...)
	result = append(result, subjectElement{
		Name:  "L",
		Value: as.config.Locality,
	}, subjectElement{
		Name:  "CN",
		Value: as.config.CommonName,
	})

	return result
}

func (as *cmpMTLSAccessStrategyExecutor) stringSliceToSubjectElementSlice(name string, values []string) []subjectElement {
	result := make([]subjectElement, 0, len(values))
	for _, val := range values {
		result = append(result, subjectElement{
			Name:  name,
			Value: val,
		})
	}
	return result
}

func (as *cmpMTLSAccessStrategyExecutor) stringSliceOrNil(str string) []string {
	if len(str) == 0 {
		return nil
	}
	return []string{str}
}
