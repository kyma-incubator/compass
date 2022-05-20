package tenantfetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

const (
	AccountCreationEventType = "create"
	AccountDeletionEventType = "delete"
	AccountUpdateEventType   = "update"

	SubaccountCreationEventType = "create_subaccount"
	SubaccountDeletionEventType = "delete_subaccount"
	SubaccountUpdateEventType   = "update_subaccount"
	SubaccountMoveEventType     = "move_subaccount"
)

type Handler struct {
	mutex            sync.Mutex
	mockedEvents     map[string][][]byte
	tenantOnDemandID string
}

func NewHandler(tenantOnDemandID string) *Handler {
	return &Handler{
		mutex:            sync.Mutex{},
		mockedEvents:     make(map[string][][]byte),
		tenantOnDemandID: tenantOnDemandID,
	}
}

func (s *Handler) HandleConfigure(eventType string) func(rw http.ResponseWriter, req *http.Request) {
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
		eventsPages, found := s.mockedEvents[eventType]
		if !found {
			eventsPages = make([][]byte, 0)
		}
		eventsPages = append(eventsPages, bodyBytes)
		s.mockedEvents[eventType] = eventsPages
		rw.WriteHeader(http.StatusOK)
		log.C(req.Context()).Infof("Tenant fetcher handler for type %s configured successfully", eventType)
	}
}

func (s *Handler) HandleFunc(eventType string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		resp := []byte("[]")
		if isSpecificSubaccountBeingFetched(req, eventType) {
			resp = getMockEventForSubaccount(req, s.tenantOnDemandID)
		} else if events, found := s.mockedEvents[eventType]; found {
			resp = events[0]
			events = events[1:]
			s.mockedEvents[eventType] = events
		}
		s.mutex.Lock()
		defer s.mutex.Unlock()
		_, err := rw.Write(resp)
		if err != nil {
			httphelpers.WriteError(rw, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		}
	}
}

func (s *Handler) HandleReset(eventType string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		log.C(req.Context()).Infof("Received a reset call for type %s. TenantFetcher queue will be emptied...", eventType)
		delete(s.mockedEvents, eventType)
		rw.WriteHeader(http.StatusOK)
	}
}

func isSpecificSubaccountBeingFetched(req *http.Request, eventType string) bool {
	entityIdParam := req.URL.Query().Get("entityId")
	return entityIdParam != "" && eventType == SubaccountCreationEventType
}

func getMockEventForSubaccount(req *http.Request, tenantOnDemandID string) []byte {
	subaccountID := req.URL.Query().Get("entityId")
	mockSubaccountEventPattern := `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"subdomain": "%s",
		"parentGuid": "%s",
		"region": "%s"
	},
	"globalAccountGUID": "%s",
	"type": "Subaccount"
}`
	emptyTenantProviderResponse := `
{
	"total": 0,
	"totalPages": 0,
	"pageNum": 0,
	"morePages": false,
	"events": []
}`
	if subaccountID == tenantOnDemandID {
		mockedEvent := fmt.Sprintf(mockSubaccountEventPattern, subaccountID, "Subaccount on demand", "subdomain", "SubaacountOnDemandParent", "region", "SubaacountOnDemandParent")
		return []byte(mockedEvent)
	}
	return []byte(emptyTenantProviderResponse)
}
