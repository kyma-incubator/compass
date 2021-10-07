package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"go.mozilla.org/pkcs7"
)

type CsrResponse struct {
	CrtResponse CrtResponse `json:"certificateChain"`
}

type CrtResponse struct {
	Crt string `json:"value"`
}

type subjectElement struct {
	Name  string `json:"shortName"`
	Value string `json:"value"`
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
	authorization := r.Header.Get("authorization")
	if len(authorization) == 0 || !strings.HasPrefix(authorization, "Bearer ") {
		httphelpers.WriteError(writer, errors.New("authorization header is required"), http.StatusBadRequest)
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

	tenant := r.Header.Get("Tenant")
	if len(tenant) == 0 { // If tenant is not provided in header, search it in the OU in the body (imitating the cert svc interface)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
			return
		}

		if len(body) > 0 {
			subjectElements := gjson.GetBytes(body, "csr.replace.subject")
			if subjectElements.Exists() {
				var subjectElementsSlice []subjectElement
				err = json.Unmarshal([]byte(subjectElements.String()), &subjectElementsSlice)
				if err != nil {
					log.C(r.Context()).WithError(err).Infof("Cannot json unmarshalling the request body. Error: %s", err)
				}
				for _, se := range subjectElementsSlice {
					if se.Name == "OU" {
						if _, err := uuid.Parse(se.Value); err == nil {
							tenant = se.Value
						}
					}
				}
			}
		}
	}

	if len(tenant) == 0 {
		httphelpers.WriteError(writer, errors.New("tenant is required"), http.StatusBadRequest)
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
