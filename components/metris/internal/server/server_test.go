package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

var certPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDCTCCAfGgAwIBAgIUaEgesRqz/TR4Qr18byordXqUlm0wDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJMTI3LjAuMC4xMB4XDTIwMDQxNzAxMTczNloXDTMwMDQx
NTAxMTczNlowFDESMBAGA1UEAwwJMTI3LjAuMC4xMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAo5pb4goXNzg4aq98yu8tr4CkB5oNQTtpANWdtuhxGsiF
MToJIxrdlEQEZZ5FBYZ7fePLwWA+zcouoIy39o8mj1Kr2s5/k/6VNiOUl8yH7bXC
W6Kcrfp2Ez66xRAwziCwOAfzVK6qTbO7CPWmS+y/WWvZ62bIZdasRe8LkHsr9NrC
a7qsn1b2AqAdeEe5o9o4RKtHX6pAVFIEHIPwF9wAoR3g87MV4cTjj0A94PpBXYkD
Ord6dgGK5ZcKD26tUFxRoKCXnPAG2KZx/1jWP5Dd6UMkEpbOVf7TFL7G6nyK1hAK
jdqfPbqd5PDtHbcHPTKrp0F2wAXfXd0QR/PQaTGM0QIDAQABo1MwUTAdBgNVHQ4E
FgQUal6x/SKO5i35lsVLyNVtJMqbNRwwHwYDVR0jBBgwFoAUal6x/SKO5i35lsVL
yNVtJMqbNRwwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAe9Ys
8m6yta4hAE+UspNVqGdqFaXrVlrgsh/bR3EPRp+Cd8y2FMPmMuVHLEb+JQ3kOVEH
5ZyAVRcharmS9GQj/NxIdLjufVocOUwtLHCr/sDduZd8njixrag93jfs+DChj+L7
nNMy2gkp13RWxYN/v7+xFbwXvD4d5hJDVnkwnAob5uBM5sk2HrHJe2DNocJFIAJI
MpYRYhnVoaZi5fjS0eaNqn480nYuqlUyJDEGeEOx4lmso9TdANZL7jHMP1OnZA1W
ggUtKaQ75uGvhNM/sS+vocVSSsC/gGwmYwAOPIkpkSNfUdAzXZpHao5eEMKLMc6n
TXTMLaTfh55GXhxYWA==
-----END CERTIFICATE-----`)

var keyPEM = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAo5pb4goXNzg4aq98yu8tr4CkB5oNQTtpANWdtuhxGsiFMToJ
IxrdlEQEZZ5FBYZ7fePLwWA+zcouoIy39o8mj1Kr2s5/k/6VNiOUl8yH7bXCW6Kc
rfp2Ez66xRAwziCwOAfzVK6qTbO7CPWmS+y/WWvZ62bIZdasRe8LkHsr9NrCa7qs
n1b2AqAdeEe5o9o4RKtHX6pAVFIEHIPwF9wAoR3g87MV4cTjj0A94PpBXYkDOrd6
dgGK5ZcKD26tUFxRoKCXnPAG2KZx/1jWP5Dd6UMkEpbOVf7TFL7G6nyK1hAKjdqf
Pbqd5PDtHbcHPTKrp0F2wAXfXd0QR/PQaTGM0QIDAQABAoIBAHyDQCbacyzNlvJr
ONjiC60khLJcZnYdFx9RdMT+RwaRUf1TJB0Zl0X+NjJ4sCMyZM49DpfP/jx1AeOi
9WerLzepPa47txEVks4vainVuiYKTS+cpQ4sLq0a8t6EIgDfE/6w3lq2zFpyzYhW
HQhv2HngAWyNPztylI5tDioZ9CxXElFqTcN7OFTkzWLvXA8nRgi8zVfitskghg9c
O0EXLYeHh3W7s/C2DmTTlTgSZuwonYcfmgSCCAr4pSSC6vrcbQ2+R2emTgugciWY
YGIMNFmJxG0KwsuRw6BQC8MZ8FP2f6s9sXJlqutAXV5Buzz5nHJjIAJoipPcQXKb
0S4+QKECgYEAz68WEsIPNrClwMjrYKlkOIo3UR1aZPLX7QhlFhHbvq5Zuq5lkUJP
opoPli6Qk9H8PIj2ouchV7s8sEGla+ZjgSGzTvW2mEIK1TUb0Wq/kPgkmKfdQL9C
b/9EgrPiGRBikx1lBBja1wyl7eFjpTsHsAxZpCH6rboMsXhfe1SLKG0CgYEAyan3
uKkeVs+q617ZH9cpsVgpLkBXO+yiKKcn0fVmdAcgartqrseAgkJspM3P6aTi7Ao4
0GR9vBXz57EQcjxTiZA/WNIvX7GvPYpYlKR3WI2O2srXMwoFDrRu/BsQT4SqMaNo
eTe0a7JjR+tUL6T/e4cy6NVUFLigjy+Ckyxyf3UCgYBFKmhNgveSHS52j/Nj08Ye
1lkp2H68U+v5cuxHd1cZn/MeKuqEf/MJkglS2Nspf6tjdbG9+v+tuhuyD4rJ7oaB
APo4d7iB6Ky26OME0WpPG3UEqhMTdx7nMbpdVQ8djclmeUmlHan0KLAyEvgelRQw
W0yXTwGvTeDgUdhquHNH2QKBgQCytKgNP/DexRCVm2uVR7puqc10axfixoO8usQS
zwCHeXlEm+iiEbDTvcFBGhFQ3wkoWraWFTdG4b1OaB5G3Sa6FNXOBBRvHKpKQrrU
nhoUov0g7fdeB1cL/OEND36YuNuJOWFvaem8Nky8gtILlo/AC8MViVYFNscxm8x+
VzjvsQKBgDEKHUeHYo5Yf/fomBGFbqOXi4Idpq2vE+4X+3P0SQcnrfp4wDLXHBp5
F9tLfqf1VyOeX2JirKXo+vkVWddXySAUd1GIMa0A485ts8RbpquhmVx7va7o7CDI
5DQ+iQKYuVlwVwGjSZTA80sJqSwgke47LMmS1Syq/8O/J3NFrDe8
-----END RSA PRIVATE KEY-----`)

func newTestServer(c Config) (testserver *httptest.Server, server *Server, err error) {
	server, err = NewServer(c, zap.NewNop().Sugar())
	if err != nil {
		return
	}

	testserver = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.Server.Handler.ServeHTTP(w, r)
	}))

	return
}

func TestServer(t *testing.T) {
	tests := []struct {
		name    string
		cert    []byte
		key     []byte
		wantErr bool
	}{
		{
			name:    "HTTP",
			cert:    nil,
			key:     nil,
			wantErr: false,
		},
		{
			name:    "TLS",
			cert:    certPEM,
			key:     keyPEM,
			wantErr: false,
		},
		{
			name:    "TLS Fail",
			cert:    certPEM,
			key:     []byte("bad key pem data"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // pin!

		t.Run(tt.name, func(t *testing.T) {
			conf := Config{}
			if tt.cert != nil && tt.key != nil {
				cert, err := ioutil.TempFile("", "cert")
				if err != nil {
					t.Fatal(err)
				}

				defer func() {
					err = os.Remove(cert.Name())
					if err != nil {
						t.Errorf("error removing cert file %v", err)
					}
				}()

				if _, err = cert.Write(tt.cert); err != nil {
					t.Fatal(err)
				}

				key, err := ioutil.TempFile("", "key")
				if err != nil {
					t.Fatal(err)
				}

				defer func() {
					err = os.Remove(key.Name())
					if err != nil {
						t.Errorf("error removing key file %v", err)
					}
				}()

				if _, err = key.Write(tt.key); err != nil {
					t.Fatal(err)
				}

				conf.TLSCertFile = cert.Name()
				conf.TLSKeyFile = key.Name()
			}

			_, _, err := newTestServer(conf)

			if !tt.wantErr && err != nil {
				t.Fatalf("expected success, got error: %s", err)
			}

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got success: %s", err)
			}
		})
	}
}

// func TestServerStartStop(t *testing.T) {
// 	_, server, err := newTestServer(Config{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(1)

// 	ctx, cancel := context.WithCancel(context.TODO())
// 	cancel()
// 	server.Start(ctx, &wg)

// 	if ctx.Err() != context.Canceled {
// 		t.Fatal(ctx.Err())
// 	}
// 	wg.Wait()
// }

func TestHealthHandler(t *testing.T) {
	httpServer, server, err := newTestServer(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer httpServer.Close()

	rr := httptest.NewRecorder()

	tests := []struct {
		name         string
		healthy      int32
		expectedCode int
	}{
		{
			name:         "Healthy",
			healthy:      1,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not Healthy",
			healthy:      0,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt // pin!

		t.Run(tt.name, func(t *testing.T) {
			atomic.StoreInt32(&server.Healthy, tt.healthy)
			server.Server.Handler.ServeHTTP(rr, httptest.NewRequest("GET", "/healthz", nil))

			if rr.Code != tt.expectedCode {
				t.Errorf("expected %d got %d", tt.expectedCode, rr.Code)
			}
		})
	}
}
