package ord_aggregator

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/form3tech-oss/jwt-go"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type ordHandler struct {
	mutex     sync.RWMutex
	secured   bool
	username  string
	password  string
	token     string
	publicKey *rsa.PublicKey
}

func NewORDHandler() *ordHandler {
	return &ordHandler{
		mutex:   sync.RWMutex{},
		secured: false,
	}
}

func (oh *ordHandler) SetPublicKey(publicKey *rsa.PublicKey) {
	oh.publicKey = publicKey
}

func (oh *ordHandler) HandleFuncOrdConfig(accessStrategy string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		oh.mutex.RLock()
		defer oh.mutex.RUnlock()

		authorizationHeader := req.Header.Get("Authorization")
		if oh.secured {
			username, password, exist := req.BasicAuth()
			if !exist {
				if authorizationHeader == "" {
					httphelpers.WriteError(rw, errors.New("missing Authorization header"), http.StatusUnauthorized)
				}
			}

			validCredentials := (username == oh.username && password == oh.password) || oh.isValidToken(authorizationHeader)

			if !validCredentials {
				httphelpers.WriteError(rw, errors.New("invalid credentials"), http.StatusUnauthorized)
			}
		}

		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(fmt.Sprintf(ordConfig, accessStrategy)))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func (oh *ordHandler) HandleFuncOrdDocument(serverPort int) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(fmt.Sprintf(ordDocument, serverPort)))
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func (oh *ordHandler) HandleFuncOrdConfigSecurity(rw http.ResponseWriter, req *http.Request) {
	oh.mutex.Lock()
	defer oh.mutex.Unlock()

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while reading request body"), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Printf("Could not close request body: %s", err)
		}
	}()

	var result struct {
		Enabled  bool   `json:"enabled"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	oh.secured = result.Enabled
	oh.username = result.Username
	oh.password = result.Password

	log.Println(fmt.Printf("Configured secured for ORD Config handler: %+v\n", result))

	rw.WriteHeader(http.StatusOK)
}

func (oh *ordHandler) isValidToken(authorizationHeader string) bool {
	if strings.Index(authorizationHeader, "Bearer") == -1 {
		return false
	}

	token := strings.TrimPrefix(authorizationHeader, "Bearer ")

	if _, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return oh.publicKey, nil
	}); err != nil {
		log.Printf("Could not validate request token: %s\n", err)
		return false

	}

	return true
}
