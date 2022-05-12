package tenantfetcher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type Handler struct {
	mutex        sync.Mutex
	mockedEvents map[string][][]byte
}

func NewHandler() *Handler {
	return &Handler{
		mutex:        sync.Mutex{},
		mockedEvents: make(map[string][][]byte),
	}
}

func (s *Handler) HandleConfigure(typee string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while reading request body"), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := req.Body.Close(); err != nil {
				log.C(req.Context()).Errorf("Could not close request body: %s", err)
			}
		}()

		var result interface{}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
			return
		}
		eventsPages, found := s.mockedEvents[typee]
		if !found {
			eventsPages = make([][]byte, 0)
		}
		eventsPages = append(eventsPages, bodyBytes)
		s.mockedEvents[typee] = eventsPages
		rw.WriteHeader(http.StatusOK)
		log.C(req.Context()).Infof("Tenant fetcher handler for type %s configured successfully", typee)
	}
}

func (s *Handler) HandleFunc(typee string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		resp := []byte("[]")
		if events, found := s.mockedEvents[typee]; found && len(events) > 0 {
			resp = events[0]
			events = events[1:]
			s.mockedEvents[typee] = events
		}
		s.mutex.Lock()
		defer s.mutex.Unlock()
		_, err := rw.Write(resp)
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func (s *Handler) HandleReset(typee string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		log.C(req.Context()).Infof("Received a reset call for type %s. TenantFetcher queue will be emptied...", typee)
		delete(s.mockedEvents, typee)
		rw.WriteHeader(http.StatusOK)
	}
}
