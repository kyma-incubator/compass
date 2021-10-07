package ord_aggregator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type ordHandler struct {
	mutex    sync.RWMutex
	secured  bool
	username string
	password string
	token    string
}

func NewORDHandler() *ordHandler {
	return &ordHandler{
		mutex:   sync.RWMutex{},
		secured: false,
	}
}

func (oh *ordHandler) HandleFuncOrdConfig(rw http.ResponseWriter, req *http.Request) {
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

		validCredentials := (username == oh.username && password == oh.password) || (authorizationHeader == oh.token)

		if !validCredentials {
			httphelpers.WriteError(rw, errors.New("invalid credentials"), http.StatusUnauthorized)
		}
	}

	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(ordConfig))
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
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
		Token    string `json:"token"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	oh.secured = result.Enabled
	oh.username = result.Username
	oh.password = result.Password
	oh.token = result.Token

	log.Println(fmt.Printf("Configured secured for ORD Config handler: %+v\n", result))

	rw.WriteHeader(http.StatusOK)
}

func (oh *ordHandler) HandleFuncOrdConfigSecurityToken(rw http.ResponseWriter, req *http.Request) {
	oh.mutex.Lock()
	defer oh.mutex.Unlock()

	bodyBytes, err := ioutil.ReadAll(req.Body)
	body := string(bodyBytes)

	clientID, clientSecret, exists := req.BasicAuth()
	if !exists {
		if !strings.Contains(body, "client_id") && !strings.Contains(body, "client_secret") {
			httphelpers.WriteError(rw, errors.New("missing Authorization header"), http.StatusUnauthorized)
		}

		split := strings.Split(body, "&")
		clientID = strings.Split(split[0], "=")[1]
		clientSecret = strings.Split(split[1], "=")[1]
	}

	validCredentials := clientID == oh.username && clientSecret == oh.password

	if !validCredentials {
		httphelpers.WriteError(rw, errors.New("invalid credentials"), http.StatusUnauthorized)
	}

	bytes, err := json.Marshal(oauth.TokenResponse{
		AccessToken: oh.token,
	})
	if err != nil {
		httphelpers.WriteError(rw, errors.New("error preparing token"), http.StatusInternalServerError)
	}

	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(bytes)
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
}

func (oh *ordHandler) HandleFuncOrdDocument(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(ordDocument))
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
}
