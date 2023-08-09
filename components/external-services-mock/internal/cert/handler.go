package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"go.mozilla.org/pkcs7"
)

type CsrResponse struct {
	CrtResponse CrtResponse `json:"certificate-response"`
}

type CrtResponse struct {
	Crt string `json:"value"`
}

type handler struct {
	CACert string
	CAKey  string
}

func NewHandler(CACert, CAKey string) *handler {
	return &handler{
		CACert: CACert,
		CAKey:  CAKey,
	}
}

func (h *handler) Generate(writer http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get(httphelpers.AuthorizationHeaderKey)

	if len(authorization) == 0 {
		httphelpers.WriteError(writer, errors.New("authorization header is required"), http.StatusBadRequest)
		return
	}

	token := strings.TrimPrefix(authorization, "Bearer ")
	if !strings.HasPrefix(authorization, "Bearer ") || len(token) == 0 {
		httphelpers.WriteError(writer, errors.New("token value is required"), http.StatusBadRequest)
		return
	}

	// Parse the CA cert
	pemBlock, _ := pem.Decode([]byte(h.CACert))

	caCRT, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}

	// Parse the CA key
	keyPemBlock, _ := pem.Decode([]byte(h.CAKey))

	caPrivateKey, err := x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
	if err != nil {
		caPrivateKeyPKCS8, err := x509.ParsePKCS8PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			httphelpers.WriteError(writer, err, http.StatusInternalServerError)
			return
		}
		var ok bool
		caPrivateKey, ok = caPrivateKeyPKCS8.(*rsa.PrivateKey)
		if !ok {
			httphelpers.WriteError(writer, errors.New("unknown CA key type"), http.StatusBadRequest)
			return
		}
	}

	// Parse CSR
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	csr := gjson.GetBytes(body, "certificate-signing-request.value")
	if !csr.Exists() {
		httphelpers.WriteError(writer, errors.Wrap(err, "missing certificate-signing-request.value in request body"), http.StatusBadRequest)
		return
	}

	csrPemBlock, _ := pem.Decode([]byte(csr.String()))
	if csrPemBlock == nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while decoding csr pem block"), http.StatusInternalServerError)
		return
	}

	clientCSR, err := x509.ParseCertificateRequest(csrPemBlock.Bytes)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while parsing CSR"), http.StatusInternalServerError)
		return
	}

	err = clientCSR.CheckSignature()
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "CSR signature invalid"), http.StatusInternalServerError)
		return
	}

	tenant := r.Header.Get("Tenant")
	if len(tenant) == 0 { // If tenant is not provided in header, search it in the OU in the body (imitating the cert svc interface)
		for _, ou := range clientCSR.Subject.OrganizationalUnit {
			if _, err := uuid.Parse(ou); err == nil {
				tenant = ou
			}
		}
	}

	cn := "compass"
	if clientCSR.Subject.CommonName != "" {
		cn = clientCSR.Subject.CommonName
	}

	if len(tenant) == 0 {
		httphelpers.WriteError(writer, errors.New("tenant is required"), http.StatusBadRequest)
		return
	}

	clientCert := x509.Certificate{
		SignatureAlgorithm: clientCSR.SignatureAlgorithm,
		SerialNumber:       big.NewInt(2),
		Subject: pkix.Name{
			Country:            []string{"DE"},
			Organization:       []string{"SAP SE"},
			OrganizationalUnit: []string{"SAP Cloud Platform Clients", "Region", tenant},
			Locality:           []string{"local"},
			CommonName:         cn,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCrtRaw, err := x509.CreateCertificate(rand.Reader, &clientCert, caCRT, clientCSR.PublicKey, caPrivateKey)
	if err != nil {
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}

	encryptedCrtContent, err := pkcs7.DegenerateCertificate(clientCrtRaw)
	if err != nil {
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}

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

	writer.Header().Set(httphelpers.ContentTypeHeaderKey, "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(payload)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
}
