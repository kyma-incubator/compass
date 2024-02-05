package service_manager

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
)

const (
	// ServiceOfferingsPath is the path for managing Service Offerings
	ServiceOfferingsPath = "/v1/service_offerings"
	// ServicePlansPath is the path for managing Service Plans
	ServicePlansPath = "/v1/service_plans"
	// ServiceInstancesPath is the path for managing Service Instances
	ServiceInstancesPath = "/v1/service_instances"
	// ServiceBindingsPath is the path for managing Service Bindings
	ServiceBindingsPath = "/v1/service_bindings"

	ServiceOfferingType = "service offering"
	ServicePlanType     = "service plan"
	ServiceInstanceType = "service instance"
	ServiceBindingType  = "service binding"

	ServiceBindingIDPath  = "serviceBindingID"
	ServiceInstanceIDPath = "serviceInstanceID"
)

type Config struct {
	Path                 string `envconfig:"APP_SERVICE_MANAGER_PATH"`
	SubaccountQueryParam string `envconfig:"APP_SERVICE_MANAGER_SUBACCOUNT_QUERY_PARAM"`
	LabelsQueryParam     string `envconfig:"APP_SERVICE_MANAGER_LABELS_QUERY_PARAM"`
}

type Handler struct {
	c Config

	ServiceInstancesMap map[string]ServiceInstancesMock // subaccount to instances
	ServiceBindingsMap  map[string]ServiceBindingsMock  // subaccount to bindings
}

func NewServiceManagerHandler(c Config) *Handler {
	return &Handler{
		c: c,

		ServiceInstancesMap: make(map[string]ServiceInstancesMock),
		ServiceBindingsMap:  make(map[string]ServiceBindingsMock),
	}
}

// Service Offerings

func (h *Handler) HandleServiceOfferingsList(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	offeringsJSON, err := json.Marshal(serviceOfferingsMock)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service offerings")
		http.Error(writer, "Failed to marshal service offerings", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(offeringsJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service offerings")
		http.Error(writer, "Failed to write service offerings", http.StatusInternalServerError)
		return
	}
}

// Service Plans

func (h *Handler) HandleServicePlansList(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	plansJSON, err := json.Marshal(servicePlansMock)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service plans")
		http.Error(writer, "Failed to marshal service plans", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(plansJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service plans")
		http.Error(writer, "Failed to write service plans", http.StatusInternalServerError)
		return
	}
}

// Service Instances

func (h *Handler) HandleServiceInstancesList(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	instances := ServiceInstancesMock{}

	labels := r.URL.Query().Get(h.c.LabelsQueryParam)
	if len(labels) == 0 {
		instances = h.ServiceInstancesMap[subaccount]
	} else {
		for _, instance := range h.ServiceInstancesMap[subaccount].Items {
			if labelsAreEqual(instance, labels) {
				instances.Items = append(instances.Items, instance)
				instances.NumItems++
			}
		}
	}
	instancesJSON, err := json.Marshal(instances)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service instances")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(instancesJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service instances")
		http.Error(writer, "Failed to write service instances", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleServiceInstanceGet(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	serviceInstanceID := mux.Vars(r)[ServiceInstanceIDPath]
	if len(serviceInstanceID) == 0 {
		http.Error(writer, "Failed to get service instance id from url", http.StatusInternalServerError)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	instance := ServiceInstanceMock{}

	labels := r.URL.Query().Get(h.c.LabelsQueryParam)

	for _, i := range h.ServiceInstancesMap[subaccount].Items {
		if i.ID == serviceInstanceID && labelsAreEqual(i, labels) {
			instance = *i
		}
	}

	if instance.ID == "" {
		log.C(ctx).Error("Service instance not found")
		http.Error(writer, "Service instance not found", http.StatusNotFound)
		return
	}

	instanceJSON, err := json.Marshal(instance)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service instances")
		http.Error(writer, "Failed to marshal service instances", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(instanceJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service instances")
		http.Error(writer, "Failed to write service instances", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleServiceInstanceCreate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while reading service instance request body"), "An error occurred while reading service instance request body", correlationID, http.StatusInternalServerError)
		return
	}

	instance := ServiceInstanceMock{}
	if err = json.Unmarshal(bodyBytes, &instance); err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while unmarshalling service instance request body"), "An error occurred while unmarshalling service instance request body", correlationID, http.StatusInternalServerError)
		return
	}

	subaccountInstancesMock := h.ServiceInstancesMap[subaccount]
	instance.ID = uuid.NewString()
	subaccountInstancesMock.Items = append(subaccountInstancesMock.Items, &instance)
	subaccountInstancesMock.NumItems++
	h.ServiceInstancesMap[subaccount] = subaccountInstancesMock

	instanceJSON, err := json.Marshal(instance)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service instance")
		http.Error(writer, "Failed to marshal service instance", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(instanceJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service instance")
		http.Error(writer, "Failed to write service instance", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleServiceInstanceDelete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	serviceInstanceID := mux.Vars(r)[ServiceInstanceIDPath]
	if len(serviceInstanceID) == 0 {
		http.Error(writer, "Failed to get service instance id from url", http.StatusInternalServerError)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	subaccountInstancesMock, found := h.ServiceInstancesMap[subaccount], false

	for i, instance := range subaccountInstancesMock.Items {
		if instance.ID == serviceInstanceID {
			subaccountInstancesMock.Items = append(subaccountInstancesMock.Items[:i], subaccountInstancesMock.Items[i+1:]...)
			subaccountInstancesMock.NumItems--
			found = true
		}
	}

	if !found {
		log.C(ctx).Error("Service binding not found")
		http.Error(writer, "Service binding not found", http.StatusNotFound)
		return
	}

	h.ServiceInstancesMap[subaccount] = subaccountInstancesMock

	httputils.Respond(writer, http.StatusOK)
}

// Service Bindings

func (h *Handler) HandleServiceBindingsList(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	bindings := h.ServiceBindingsMap[subaccount]

	bindingsJSON, err := json.Marshal(bindings)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service bindings")
		http.Error(writer, "Failed to marshal service bindings", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(bindingsJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service bindings")
		http.Error(writer, "Failed to write service bindings", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleServiceBindingGet(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	serviceBindingID := vars[ServiceBindingIDPath]
	if len(serviceBindingID) == 0 {
		http.Error(writer, "Failed to get service binding id from url", http.StatusInternalServerError)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	binding := ServiceBindingMock{}
	for _, b := range h.ServiceBindingsMap[subaccount].Items {
		if b.ID == serviceBindingID {
			binding = *b
		}
	}

	if binding.ID == "" {
		log.C(ctx).Error("Service binding not found")
		http.Error(writer, "Service binding not found", http.StatusNotFound)
		return
	}

	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service binding")
		http.Error(writer, "Failed to marshal service binding", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(bindingJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service binding")
		http.Error(writer, "Failed to write service binding", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleServiceBindingCreate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while reading service binding request body"), "An error occurred while reading service binding request body", correlationID, http.StatusInternalServerError)
		return
	}

	binding := ServiceBindingMock{}
	if err = json.Unmarshal(bodyBytes, &binding); err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while unmarshalling service binding request body"), "An error occurred while unmarshalling service binding request body", correlationID, http.StatusInternalServerError)
		return
	}

	subaccountBindingsMock := h.ServiceBindingsMap[subaccount]
	binding.ID = uuid.NewString()
	subaccountBindingsMock.Items = append(subaccountBindingsMock.Items, &binding)
	subaccountBindingsMock.NumItems++
	h.ServiceBindingsMap[subaccount] = subaccountBindingsMock

	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal service binding")
		http.Error(writer, "Failed to marshal service binding", http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(bindingJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write service binding")
		http.Error(writer, "Failed to write service binding", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleServiceBindingDelete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	// Validate that there is a bearer token
	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	serviceBindingID := mux.Vars(r)[ServiceBindingIDPath]
	if len(serviceBindingID) == 0 {
		http.Error(writer, "Failed to get service binding id from url", http.StatusInternalServerError)
		return
	}

	subaccount := r.URL.Query().Get(h.c.SubaccountQueryParam)
	if len(subaccount) == 0 {
		log.C(ctx).Error("Failed to get subaccount from query")
		http.Error(writer, "Failed to get subaccount from query", http.StatusInternalServerError)
		return
	}

	subaccountBindingsMock, found := h.ServiceBindingsMap[subaccount], false

	for i, binding := range subaccountBindingsMock.Items {
		if binding.ID == serviceBindingID {
			subaccountBindingsMock.Items = append(subaccountBindingsMock.Items[:i], subaccountBindingsMock.Items[i+1:]...)
			subaccountBindingsMock.NumItems--
			found = true
		}
	}

	if !found {
		log.C(ctx).Error("Service binding not found")
		http.Error(writer, "Service binding not found", http.StatusNotFound)
		return
	}

	h.ServiceBindingsMap[subaccount] = subaccountBindingsMock

	httputils.Respond(writer, http.StatusOK)
}

// ServiceOfferingsMock represents a collection of Service Offering Mocks
type ServiceOfferingsMock struct {
	NumItems int                    `json:"num_items"`
	Items    []*ServiceOfferingMock `json:"items"`
}

// ServiceOfferingMock represents a Service Offering mock
type ServiceOfferingMock struct {
	ID          string `json:"id"`
	CatalogName string `json:"catalog_name"`
}

// ServicePlanMock represents a Service Plan
type ServicePlanMock struct {
	ID                string `json:"id"`
	CatalogName       string `json:"catalog_name"`
	ServiceOfferingId string `json:"service_offering_id"`
}

// ServicePlansMock represents a collection of Service Plan
type ServicePlansMock struct {
	NumItems int                `json:"num_items"`
	Items    []*ServicePlanMock `json:"items"`
}

var (
	serviceOfferingsMock = ServiceOfferingsMock{
		NumItems: 2,
		Items: []*ServiceOfferingMock{
			{
				ID:          "first-service-offering-id",
				CatalogName: "first-service-offering-test",
			},
			{
				ID:          "second-service-offering-id",
				CatalogName: "second-service-offering-test",
			},
		},
	}

	servicePlansMock = ServicePlansMock{
		NumItems: 2,
		Items: []*ServicePlanMock{
			{
				ID:                "1",
				CatalogName:       "first-catalog-name",
				ServiceOfferingId: "first-service-offering-id",
			},
			{
				ID:                "2",
				CatalogName:       "second-catalog-name",
				ServiceOfferingId: "second-service-offering-id",
			},
		},
	}
)

// ServiceInstancesMock represents a collection of Service Instance Mock
type ServiceInstancesMock struct {
	NumItems int                    `json:"num_items"`
	Items    []*ServiceInstanceMock `json:"items"`
}

// ServiceInstanceMock represents a Service Instance
type ServiceInstanceMock struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	ServicePlanID string              `json:"service_plan_id"`
	PlatformID    string              `json:"platform_id"`
	Labels        map[string][]string `json:"labels,omitempty"`
}

// ServiceBindingMock represents a Service Binding Mock
type ServiceBindingMock struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	ServiceInstanceID string          `json:"service_instance_id"`
	Credentials       json.RawMessage `json:"credentials"`
}

// ServiceBindingsMock represents a collection of Service Binding Mocks
type ServiceBindingsMock struct {
	NumItems int                   `json:"num_items"`
	Items    []*ServiceBindingMock `json:"items"`
}

func validateAuthorization(ctx context.Context, r *http.Request) error {
	log.C(ctx).Info("Validating authorization header...")
	authorizationHeaderValue := r.Header.Get(httphelpers.AuthorizationHeaderKey)

	if authorizationHeaderValue == "" {
		return errors.New("Missing authorization header")
	}

	tokenValue := strings.TrimSpace(strings.TrimPrefix(authorizationHeaderValue, "Bearer "))
	if tokenValue == "" {
		return errors.New("The token value cannot be empty")
	}

	return nil
}

func labelsAreEqual(instance *ServiceInstanceMock, labels string) bool {
	if len(labels) == 0 {
		return true
	}

	marshalledInstanceLabels, err := json.Marshal(instance.Labels)
	if err != nil {
		return false
	}

	return string(marshalledInstanceLabels) == labels
}
