package systemfetcher

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type SystemFetcherHandler struct {
	mutex         sync.Mutex
	defaulTenant  string
	mockedSystems [][]byte
}

func NewSystemFetcherHandler(defaultTenant string) *SystemFetcherHandler {
	return &SystemFetcherHandler{
		mutex:         sync.Mutex{},
		defaulTenant:  defaultTenant,
		mockedSystems: make([][]byte, 0),
	}
}

func (s *SystemFetcherHandler) HandleConfigure(rw http.ResponseWriter, req *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
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

	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	s.mockedSystems = append(s.mockedSystems, bodyBytes)
	rw.WriteHeader(http.StatusOK)
}

func (s *SystemFetcherHandler) HandleFunc(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	resp := []byte("[]")
	if len(s.mockedSystems) > 0 {
		resp = s.mockedSystems[0]
		s.mockedSystems = s.mockedSystems[1:]
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := rw.Write(resp)
	if err != nil {
		httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
}

func (s *SystemFetcherHandler) HandleReset(rw http.ResponseWriter, _ *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	log.Println("Recieved a reset call.SystemFetcher queue will be emptied...")
	s.mockedSystems = make([][]byte, 0)
	rw.WriteHeader(http.StatusOK)
}
