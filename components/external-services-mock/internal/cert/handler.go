package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"go.mozilla.org/pkcs7"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type CsrResponse struct {
	CrtResponse CrtResponse `json:"certificateChain"`
}

type CrtResponse struct {
	Crt string `json:"value"`
}

type handler struct {
	CACert []byte
	CAKey  []byte
}

func NewHandler(CACert, CAKey []byte) *handler {
	return &handler{
		CACert: CACert,
		CAKey:  CAKey,
	}
}

func (h *handler) Generate(writer http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("authorization")
	if len(authorization) == 0 || !strings.HasPrefix(authorization, "Bearer ") {
		httphelpers.WriteError(writer, errors.New("authorization header is required"), http.StatusBadRequest)
		return
	}

	// Parse the CA cert
	pemBlock, _ := pem.Decode(h.CACert)

	caCRT, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}

	// Parse the CA key
	keyPemBlock, _ := pem.Decode(h.CAKey)

	caPrivateKey, err := x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
	if err != nil {
		caPrivateKeyPKCS8, err := x509.ParsePKCS8PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			httphelpers.WriteError(writer, err, http.StatusInternalServerError)
			return
		}
		caPrivateKey = caPrivateKeyPKCS8.(*rsa.PrivateKey)
	}

	tenant := r.Header.Get("Tenant")
	if len(tenant) == 0 {
		httphelpers.WriteError(writer, errors.New("tenant header is required"), http.StatusBadRequest)
		return
	}

	clientCert := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Country:            []string{"DE"},
			Organization:       []string{"SAP SE"},
			OrganizationalUnit: []string{"SAP Cloud Platform Clients", "Region", tenant},
			Locality:           []string{"local"},
			CommonName:         "compass",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}

	clientCrtRaw, err := x509.CreateCertificate(rand.Reader, &clientCert, caCRT, &clientKey.PublicKey, caPrivateKey)
	if err != nil {
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}

	encryptedCrtContent, err := pkcs7.Encrypt(clientCrtRaw, nil)
	crt := pem.EncodeToMemory(&pem.Block{
		Type: "PKCS7", Bytes: encryptedCrtContent,
	})

	response := CsrResponse{
		CrtResponse: CrtResponse{
			Crt: string(crt),
		},
	}
	payload, err := json.Marshal(response)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(payload)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
}