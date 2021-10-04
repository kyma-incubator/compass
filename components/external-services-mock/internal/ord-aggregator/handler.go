package ord_aggregator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type ordHandler struct {
	mutex    sync.Mutex
	secured  bool
	username string
	password string
}

func NewORDHandler() *ordHandler {
	return &ordHandler{
		mutex:   sync.Mutex{},
		secured: false,
	}
}

func (oh *ordHandler) HandleFuncOrdConfig(rw http.ResponseWriter, req *http.Request) {
	oh.mutex.Lock()
	defer oh.mutex.Unlock()

	fmt.Println("Inside ORD Config")
	if oh.secured {
		username, password, exist := req.BasicAuth()
		fmt.Printf("Actual Username: %s, Password: %s", username, password)
		fmt.Printf("Expected Username: %s, Password: %s", oh.username, oh.password)
		if !exist {
			httphelpers.WriteError(rw, errors.New("missing Authorization header"), http.StatusUnauthorized)
		}

		if username != oh.username || password == oh.password {
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

func (oh *ordHandler) HandleFuncOrdDocument(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(ordDocument))
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
}
