package cert_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/cert"
	"github.com/stretchr/testify/require"
)

type csrRequest struct {
	Csr csrPayload `json:"certificate-signing-request"`
}

type csrPayload struct {
	Value string `json:"value"`
}

func TestHandler_Generate(t *testing.T) {
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{"DE"},
			Organization:       []string{"SAP SE"},
			OrganizationalUnit: []string{"OU"},
			Locality:           []string{"local"},
			CommonName:         "compass test",
		},
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, clientKey)
	require.NoError(t, err)

	pemEncodedCSR := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr,
	})

	data, err := json.Marshal(&csrRequest{csrPayload{Value: string(pemEncodedCSR)}})
	require.NoError(t, err)

	//GIVEN
	req := httptest.NewRequest(http.MethodPost, "http://target.com/cert", bytes.NewBuffer(data))
	req.Header.Set(httphelpers.AuthorizationHeaderKey, "Bearer test-tkn")
	req.Header.Set("tenant", "tnt")

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	caCert := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Country:            []string{"DE"},
			Organization:       []string{"SAP SE"},
			OrganizationalUnit: []string{"OU"},
			Locality:           []string{"local"},
			CommonName:         "ca compass test",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	caCrtRaw, err := x509.CreateCertificate(rand.Reader, &caCert, &caCert, &key.PublicKey, key)
	require.NoError(t, err)

	caCrtPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCrtRaw})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	h := cert.NewHandler(string(caCrtPem), string(keyPem))
	r := httptest.NewRecorder()

	//WHEN
	h.Generate(r, req)
	resp := r.Result()

	//THEN
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var response cert.CsrResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.NotEmpty(t, response.CrtResponse.Crt)
}
