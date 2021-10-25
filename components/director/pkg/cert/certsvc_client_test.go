package cert_test

import (
	"bytes"
	"context"
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
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.mozilla.org/pkcs7"
)

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

type csrResponse struct {
	CrtResponse crtResponse `json:"certificateChain"`
}

type crtResponse struct {
	Crt string `json:"value"`
}

func TestIssueClientCert(t *testing.T) {
	// Issue a mock test client cert
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tenant := "test-tenant"
	testCert := x509.Certificate{
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

	testCrtRaw, err := x509.CreateCertificate(rand.Reader, &testCert, &testCert, &key.PublicKey, key)
	require.NoError(t, err)

	encryptedCrtContent, err := pkcs7.DegenerateCertificate(testCrtRaw)
	require.NoError(t, err)

	crt := pem.EncodeToMemory(&pem.Block{
		Type: "PKCS7", Bytes: encryptedCrtContent,
	})

	// Issue a mock test client cert with invalid Locality
	wrongLocalityTestCert := testCert
	wrongLocalityTestCert.Subject.Locality = []string{"not-local"}
	wrongLocalityTestCrtRaw, err := x509.CreateCertificate(rand.Reader, &wrongLocalityTestCert, &wrongLocalityTestCert, &key.PublicKey, key)
	require.NoError(t, err)

	encryptedWrongLocalityCrtContent, err := pkcs7.DegenerateCertificate(wrongLocalityTestCrtRaw)
	require.NoError(t, err)

	wrongLocalityCrt := pem.EncodeToMemory(&pem.Block{
		Type: "PKCS7", Bytes: encryptedWrongLocalityCrtContent,
	})

	oAuthURL := "http://oauth"
	clientSecret := "client_secret"
	clientID := "client_id"

	csrEndpoint := "http://csr"
	policy := "sap-cloud-platform-clients"

	fullConfig := cert.CertSvcConfig{
		SubjectPattern: "C=DE, O=SAP SE, OU=SAP Cloud Platform Clients, OU=Canary, OU=209208f2-331f-4fc4-8565-1cb7b260d13f, L=%s, CN=%s",
		CommonName:     "CMP",
		Locality:       "Local",
		Policy:         policy,
		CSREndpoint:    csrEndpoint,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		OAuthURL:       oAuthURL,
		IssuerLocality: "local",
		RetryAttempts:  2,
		RetryDelay:     100 * time.Millisecond,
	}

	successfulClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		body := string(bodyBytes)

		var data []byte

		if strings.Contains(req.URL.String(), oAuthURL) && strings.Contains(body, clientSecret) && strings.Contains(body, clientID) {
			data, err = json.Marshal(struct {
				AccessToken string `json:"access_token"`
			}{
				AccessToken: "test-tkn",
			})
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), csrEndpoint) && strings.Contains(body, policy) {
			token := req.Header.Get("Authorization")
			require.Equal(t, "Bearer test-tkn", token)
			data, err = json.Marshal(csrResponse{
				CrtResponse: crtResponse{
					Crt: string(crt),
				},
			})
			require.NoError(t, err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
		}, nil
	})

	var isCertWithWrongLocality = true
	successfulClientAfterOneRetry := newTestClient(func(req *http.Request) (*http.Response, error) {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		body := string(bodyBytes)

		var data []byte

		if strings.Contains(req.URL.String(), oAuthURL) && strings.Contains(body, clientSecret) && strings.Contains(body, clientID) {
			data, err = json.Marshal(struct {
				AccessToken string `json:"access_token"`
			}{
				AccessToken: "test-tkn",
			})
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), csrEndpoint) && strings.Contains(body, policy) && isCertWithWrongLocality {
			token := req.Header.Get("Authorization")
			require.Equal(t, "Bearer test-tkn", token)
			data, err = json.Marshal(csrResponse{
				CrtResponse: crtResponse{
					Crt: string(wrongLocalityCrt),
				},
			})
			require.NoError(t, err)
			isCertWithWrongLocality = !isCertWithWrongLocality
		} else if strings.Contains(req.URL.String(), csrEndpoint) && strings.Contains(body, policy) && !isCertWithWrongLocality {
			token := req.Header.Get("Authorization")
			require.Equal(t, "Bearer test-tkn", token)
			data, err = json.Marshal(csrResponse{
				CrtResponse: crtResponse{
					Crt: string(crt),
				},
			})
			require.NoError(t, err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
		}, nil
	})

	unsuccessfulClientAfterAllRetries := newTestClient(func(req *http.Request) (*http.Response, error) {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		body := string(bodyBytes)

		var data []byte

		if strings.Contains(req.URL.String(), oAuthURL) && strings.Contains(body, clientSecret) && strings.Contains(body, clientID) {
			data, err = json.Marshal(struct {
				AccessToken string `json:"access_token"`
			}{
				AccessToken: "test-tkn",
			})
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), csrEndpoint) && strings.Contains(body, policy) {
			token := req.Header.Get("Authorization")
			require.Equal(t, "Bearer test-tkn", token)
			data, err = json.Marshal(csrResponse{
				CrtResponse: crtResponse{
					Crt: string(wrongLocalityCrt),
				},
			})
			require.NoError(t, err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
		}, nil
	})

	assertNotCalledClient := newTestClient(func(req *http.Request) (*http.Response, error) {
		t.FailNow()
		return nil, nil
	})

	for _, testCase := range []struct {
		Name        string
		Config      cert.CertSvcConfig
		Client      *http.Client
		ExpectedErr string
	}{
		{
			Name:   "Success",
			Config: fullConfig,
			Client: successfulClient,
		},
		{
			Name:   "Success with one retry",
			Config: fullConfig,
			Client: successfulClientAfterOneRetry,
		},
		{
			Name:        "Unsuccessful after all retries are exhausted",
			Config:      fullConfig,
			Client:      unsuccessfulClientAfterAllRetries,
			ExpectedErr: "issuer locality of the client cert does not match the expected one",
		},
		{
			Name:        "Invalid config should return err",
			Client:      assertNotCalledClient,
			ExpectedErr: "invalid config",
		},
		{
			Name:   "Token request fail should return err",
			Config: fullConfig,
			Client: newTestClient(func(req *http.Request) (*http.Response, error) {
				bodyBytes, err := ioutil.ReadAll(req.Body)
				require.NoError(t, err)
				body := string(bodyBytes)

				if strings.Contains(req.URL.String(), oAuthURL) && strings.Contains(body, clientSecret) && strings.Contains(body, clientID) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       nil,
					}, nil
				}

				return nil, errors.New("should not happen")
			}),
			ExpectedErr: "cannot fetch token",
		},
		{
			Name:   "CSR signing fail should return err",
			Config: fullConfig,
			Client: newTestClient(func(req *http.Request) (*http.Response, error) {
				bodyBytes, err := ioutil.ReadAll(req.Body)
				require.NoError(t, err)
				body := string(bodyBytes)

				if strings.Contains(req.URL.String(), oAuthURL) && strings.Contains(body, clientSecret) && strings.Contains(body, clientID) {
					data, err := json.Marshal(struct {
						AccessToken string `json:"access_token"`
					}{
						AccessToken: "test-tkn",
					})
					require.NoError(t, err)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
					}, nil
				} else if strings.Contains(req.URL.String(), csrEndpoint) && strings.Contains(body, policy) {
					token := req.Header.Get("Authorization")
					require.Equal(t, "Bearer test-tkn", token)
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       nil,
					}, nil
				}

				return nil, errors.New("should not happen")
			}),
			ExpectedErr: "unexpected status code while issuing client cert",
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			client := cert.NewCertSvcClient(testCase.Client, testCase.Config)

			clientCert, err := client.IssueClientCert(context.Background())

			if len(testCase.ExpectedErr) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr)
				require.Nil(t, clientCert)
			} else {
				require.NoError(t, err)
				require.NotNil(t, clientCert)
			}
		})
	}
}
