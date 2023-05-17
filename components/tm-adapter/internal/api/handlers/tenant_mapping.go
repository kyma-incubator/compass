package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/paths"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/types"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/external_caller"
	"github.com/kyma-incubator/compass/components/tm-adapter/pkg/httputil"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

const (
	SubaccountKey     = "subaccount_id"
	LocationHeaderKey = "Location"
)

type Handler struct {
	cfg      *config.Config
	caller   *external_caller.Caller
	tenantID string
}

func NewHandler(cfg *config.Config, caller *external_caller.Caller) *Handler {
	return &Handler{
		cfg:    cfg,
		caller: caller,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	errResp := errors.Errorf("An unexpected error occurred while processing the request. X-Request-Id: %s", correlationID)

	log.C(ctx).Infof("Processing tenant mapping notification...")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("Failed to read request body: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Failed to read request body"))
		return
	}
	log.C(ctx).Infof("Tenant mapping request body: %s", reqBody)

	formationID := gjson.Get(string(reqBody), "context.btp.uclFormationId").String()
	if formationID == "" {
		log.C(ctx).Error("Failed to get the formation ID from the tenant mapping request body")
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Failed to get the formation ID from the tenant mapping request body"))
		return
	}

	var tm types.TenantMapping
	err = json.Unmarshal(reqBody, &tm)
	if err != nil {
		log.C(ctx).Errorf("Failed to unmarshal request body: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Invalid json"))
		return
	}

	if err := validate(tm); err != nil {
		log.C(ctx).Errorf("Failed to validate tenant mapping request body: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New(""))
		return
	}

	if tm.Items[0].Configuration != nil && string(tm.Items[0].Configuration) != "{}" && string(tm.Items[0].Configuration) != "\"\"" {
		log.C(ctx).Info("The configuration in the tenant mapping body is provided and no service instance/binding will be created. Returning...")
		httputil.Respond(w, http.StatusOK)
		return
	}

	log.C(ctx).Info("Creating procurement service instance")

	catalogNameProcurement := "procurement-service" // todo::: most probably should be provided as label on the runtime/app-template and will be used through TM notification body
	planNameProcurement := "apiaccess"              // todo::: most probably should be provided as label on the runtime/app-template and will be used through TM notification body
	h.tenantID = tm.ReceiverTenant.SubaccountID

	offeringIDProcurement, err := h.retrieveServiceOffering(ctx, catalogNameProcurement)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	planIDProcurement, err := h.retrieveServicePlan(ctx, planNameProcurement, offeringIDProcurement)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	svcInstanceNameProcurement := catalogNameProcurement + "-instance-" + formationID
	_, err = h.createServiceInstance(ctx, planIDProcurement, svcInstanceNameProcurement)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	log.C(ctx).Info("Creating IAS service instance and key")

	catalogNameIAS := "identity" // IAS
	planNameIAS := "application"

	offeringIDIAS, err := h.retrieveServiceOffering(ctx, catalogNameIAS)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	planIDIAS, err := h.retrieveServicePlan(ctx, planNameIAS, offeringIDIAS)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	svcInstanceNameIAS := catalogNameIAS + "-instance-" + formationID
	svcInstanceIDIAS, err := h.createServiceInstance(ctx, planIDIAS, svcInstanceNameIAS)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	// todo:: consider removing it
	//svcInstanceIDIAS, err := h.retrieveServiceInstanceByName(ctx, svcInstanceNameIAS)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}

	svcKeyNameIAS := svcInstanceNameIAS + "-key"
	serviceKeyIDIAS, err := h.createServiceKey(ctx, svcKeyNameIAS, svcInstanceIDIAS, svcInstanceNameProcurement)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	// todo:: consider removing it
	//serviceKeyIAS, err = h.retrieveServiceKeyByName(ctx, svcKeyNameIAS)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}

	serviceKeyIAS, err := h.retrieveServiceKeyByID(ctx, serviceKeyIDIAS)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	//// old response body containing service key credentials
	//responseBody := types.Response{Configuration: serviceKey.Credentials} /// todo::: consider removing it?

	if len(serviceKeyIAS.Credentials) < 0 {
		log.C(ctx).Errorf("The credentials for service key with ID: %q should not be empty", serviceKeyIAS.ID)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	data, err := h.buildTemplateData(serviceKeyIAS.Credentials, tm)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	mockURL := "https://uclmockapi.free.beeceptor.com/v1/tenantMappings/"
	req, err := http.NewRequest(http.MethodPatch, mockURL, nil)
	if err != nil {
		log.C(ctx).Error("An error occurred while creating request to the mock API")
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	log.C(ctx).Info("Calling beeceptor mock API...")
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error("An error occurred while calling beeceptor mock API")
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).Errorf("Failed to read response body from beeceptor mock API request: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Errorf("Response status code is not the exepcted one, should be: %d, got: %d", http.StatusOK, resp.StatusCode)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	t, err := template.New("").Parse(string(body))
	if err != nil {
		log.C(ctx).Errorf("An error occurred while creating template: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		log.C(ctx).Errorf("An error occurred while executing template: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	var jsonRawMsg json.RawMessage
	err = json.Unmarshal([]byte(res.String()), &jsonRawMsg)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while preparing response: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	log.C(ctx).Infof("Successfully processed tenant mapping notification")
	httputil.RespondWithBody(ctx, w, http.StatusOK, jsonRawMsg)
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func validate(tm types.TenantMapping) error {
	if len(tm.Items) != 1 {
		return errors.New("The items in the tenant mapping request body should consists of one element")
	}

	if tm.ReceiverTenant.SubaccountID == "" {
		return errors.New("The subaccount ID in the tenant mapping request body should not be empty")
	}

	return nil
}

func (h *Handler) retrieveServiceOffering(ctx context.Context, catalogName string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceOfferingsPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service offerings URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Listing service offerings...")
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read service offerings response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Failed to get service offerings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service offerings")

	var offerings types.ServiceOfferings
	err = json.Unmarshal(body, &offerings)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service offerings: %v", err)
	}

	var offeringID string
	for _, item := range offerings.Items {
		if item.CatalogName == catalogName {
			offeringID = item.ID
			break
		}
	}
	log.C(ctx).Infof("Service offering ID: %q", offeringID)

	return offeringID, nil
}

func (h *Handler) retrieveServicePlan(ctx context.Context, planName, offeringID string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServicePlansPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service plans URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Listing service plans...")
	resp, err := h.caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read service plans response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Failed to get service plans, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service plans")

	var plans types.ServicePlans
	err = json.Unmarshal(body, &plans)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service plans: %v", err)
	}

	var planID string
	for _, item := range plans.Items {
		if item.CatalogName == planName && item.ServiceOfferingId == offeringID {
			planID = item.ID
			break
		}
	}
	log.C(ctx).Infof("Service plan ID: %q", planID)

	return planID, nil
}

func (h *Handler) createServiceInstance(ctx context.Context, planID, serviceInstanceName string) (string, error) {
	siReqBody := &types.ServiceInstanceReqBody{
		Name:          serviceInstanceName,
		ServicePlanId: planID,
	}

	siReqBodyBytes, err := json.Marshal(siReqBody)
	if err != nil {
		return "", errors.Errorf("Failed to marshal service instance body: %v", err)
	}

	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceInstancesPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(siReqBodyBytes))
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Creating service instance with name: %q from plan with ID: %q", serviceInstanceName, planID)
	resp, err := h.caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read response body from service instance creation request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", errors.Errorf("Failed to create service instance, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous service instance creation...")
		opStatusPath := resp.Header.Get(LocationHeaderKey)
		if opStatusPath == "" {
			return "", errors.Errorf("The service instance operation status path from %s header should not be empty", LocationHeaderKey)
		}

		opURL, err := buildURL(h.cfg.ServiceManagerURL, opStatusPath, SubaccountKey, h.tenantID)
		if err != nil {
			return "", errors.Wrapf(err, "while building asynchronous service instance operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return "", err
		}

		ticker := time.NewTicker(3 * time.Second)
		timeout := time.After(time.Second * 15) // todo::: extract as config, valid for the ticker as well
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service instance with name: %q", serviceInstanceName)
				opResp, err := h.caller.Call(opReq)
				if err != nil {
					return "", err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := ioutil.ReadAll(opResp.Body)
				if err != nil {
					return "", errors.Errorf("Failed to read operation response body from asynchronous service instance creation request: %v", err)
				}

				if opResp.StatusCode != http.StatusOK {
					return "", errors.Errorf("Failed to get asynchronous service instance operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
				}

				var opStatus types.OperationStatus
				err = json.Unmarshal(opBody, &opStatus)
				if err != nil {
					return "", errors.Errorf("Failed to unmarshal service instance operation status: %v", err)
				}

				if opStatus.State == types.OperationStateInProgress {
					log.C(ctx).Infof("The asynchronous service instance operation state is still: %q", types.OperationStateInProgress)
					continue
				}

				if opStatus.State != types.OperationStateSucceeded {
					return "", errors.Errorf("The asynchronous service instance operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
				}

				log.C(ctx).Infof("The asynchronous operation status for service instance with name: %q finished with state: %q", serviceInstanceName, opStatus.State)
				serviceInstanceID := opStatus.ResourceID
				if serviceInstanceID == "" {
					return "", errors.New("The service instance ID could not be empty")
				}

				return serviceInstanceID, nil
			case <-timeout:
				return "", errors.New("Timeout waiting for asynchronous operation status to finish")
			}
		}
	}

	log.C(ctx).Infof("Successfully create service instance with name: %q synchronously", serviceInstanceName)
	var serviceInstance types.ServiceInstance
	err = json.Unmarshal(body, &serviceInstance)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service instance: %v", err)
	}

	serviceInstanceID := serviceInstance.ID
	if serviceInstanceID == "" {
		return "", errors.New("The service instance ID could not be empty")
	}
	log.C(ctx).Infof("Service instance ID: %q", serviceInstanceID)

	return serviceInstanceID, nil
}

// todo:: consider removing retrieveServiceInstanceByName
func (h *Handler) retrieveServiceInstanceByName(ctx context.Context, serviceInstanceName string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceInstancesPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Listing service instances...")
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read service instances response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Failed to get service offerings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service offerings")

	var instances types.ServiceInstances
	err = json.Unmarshal(body, &instances)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service instances: %v", err)
	}

	var instanceID string
	for _, item := range instances.Items {
		if item.Name == serviceInstanceName {
			instanceID = item.ID
			break
		}
	}
	log.C(ctx).Infof("Service instance ID: %q", instanceID)

	return instanceID, nil
}

// todo:: double check
func (h *Handler) retrieveServiceInstanceByID(ctx context.Context, serviceInstanceID string) (string, error) {
	svcInstancePath := paths.ServiceInstancesPath + fmt.Sprintf("/%s", serviceInstanceID)
	strURL, err := buildURL(h.cfg.ServiceManagerURL, svcInstancePath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Getting service instance by ID: %q", serviceInstanceID)
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read service instance response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Failed to get service instance, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service instance by ID: %q", serviceInstanceID)

	var instance types.ServiceInstance
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service instances: %v", err)
	}
	log.C(ctx).Infof("Service instance ID: %q", instance.ID)

	return instance.ID, nil
}

func (h *Handler) createServiceKey(ctx context.Context, serviceKeyName, serviceInstanceID, serviceInstanceNameProcurement string) (string, error) {
	iasParams := types.IASParameters{
		ConsumedServices: []types.ConsumedService{
			{
				ServiceInstanceName: serviceInstanceNameProcurement,
			},
		},
		XsuaaCrossConsumption: true,
	}

	iasParamsBytes, err := json.Marshal(iasParams)
	if err != nil {
		return "", errors.Errorf("Failed to marshal IAS parameters with procurement service details: %v", err)
	}

	serviceKeyReqBody := &types.ServiceKeyReqBody{
		Name:              serviceKeyName,
		ServiceInstanceId: serviceInstanceID,
		Parameters:        iasParamsBytes, // todo::: most probably should be provided as `parameters` label in the TM notification body - `receiverTenant.parameters`?
	}

	serviceKeyReqBodyBytes, err := json.Marshal(serviceKeyReqBody)
	if err != nil {
		return "", errors.Errorf("Failed to marshal service key body: %v", err)
	}

	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceBindingsPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service bindings URL")
	}

	log.C(ctx).Infof("Creating IAS service key for service instance with name: %q", serviceInstanceNameProcurement)
	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(serviceKeyReqBodyBytes))
	if err != nil {
		return "", err
	}

	resp, err := h.caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read response body from service key creation request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", errors.Errorf("Failed to create service key, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous service key creation...")
		opStatusPath := resp.Header.Get(LocationHeaderKey)
		if opStatusPath == "" {
			return "", errors.Errorf("The service key operation status path from %s header should not be empty", LocationHeaderKey)
		}

		opURL, err := buildURL(h.cfg.ServiceManagerURL, opStatusPath, SubaccountKey, h.tenantID)
		if err != nil {
			return "", errors.Wrapf(err, "while building asynchronous service key operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return "", err
		}

		ticker := time.NewTicker(3 * time.Second)
		timeout := time.After(time.Second * 15) // todo::: extract as config, valid for the ticker as well
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service key with name: %q", serviceKeyName)
				opResp, err := h.caller.Call(opReq)
				if err != nil {
					return "", err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := ioutil.ReadAll(opResp.Body)
				if err != nil {
					return "", errors.Errorf("Failed to read operation response body from asynchronous service key creation request: %v", err)
				}

				if opResp.StatusCode != http.StatusOK {
					return "", errors.Errorf("Failed to get asynchronous service key operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
				}

				var opStatus types.OperationStatus
				err = json.Unmarshal(opBody, &opStatus)
				if err != nil {
					return "", errors.Errorf("Failed to unmarshal service key operation status: %v", err)
				}

				if opStatus.State == types.OperationStateInProgress {
					log.C(ctx).Infof("The asynchronous service key operation state is still: %q", types.OperationStateInProgress)
					continue
				}

				if opStatus.State != types.OperationStateSucceeded {
					return "", errors.Errorf("The asynchronous service key operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
				}

				log.C(ctx).Infof("The asynchronous operation status for service key with name: %q finished with state: %q", serviceKeyName, opStatus.State)
				serviceKeyID := opStatus.ResourceID
				if serviceKeyID == "" {
					return "", errors.New("The service key ID could not be empty")
				}

				return serviceKeyID, nil
			case <-timeout:
				return "", errors.New("Timeout waiting for asynchronous operation status to finish")
			}
		}
	}

	log.C(ctx).Infof("Successfully create IAS service key for service instance with name: %q synchronously", serviceInstanceNameProcurement)
	var serviceKey types.ServiceKey
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service key: %v", err)
	}

	serviceKeyID := serviceKey.ID
	if serviceKeyID == "" {
		return "", errors.New("The service key ID could not be empty")
	}
	log.C(ctx).Infof("Service key ID: %q", serviceKeyID)

	return serviceKeyID, nil
}

// todo:: consider removing retrieveServiceKeyByName
func (h *Handler) retrieveServiceKeyByName(ctx context.Context, serviceKeyName string) (*types.ServiceKey, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceBindingsPath, SubaccountKey, h.tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building service binding URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Listing service bindings...")
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("Failed to read service binding response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed to get service bindings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service bindings")

	var svcKeys types.ServiceKeys
	err = json.Unmarshal(body, &svcKeys)
	if err != nil {
		return nil, errors.Errorf("Failed to unmarshal service keys: %v", err)
	}

	var serviceKey types.ServiceKey
	for _, item := range svcKeys.Items {
		if item.Name == serviceKeyName {
			serviceKey = item
			break
		}
	}
	log.C(ctx).Infof("Service key ID: %q", serviceKey.ID)

	return &serviceKey, nil
}

func (h *Handler) retrieveServiceKeyByID(ctx context.Context, serviceKeyID string) (*types.ServiceKey, error) {
	svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", serviceKeyID)
	strURL, err := buildURL(h.cfg.ServiceManagerURL, svcKeyPath, SubaccountKey, h.tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building service binding URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Getting service key by ID: %q", serviceKeyID)
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("Failed to read service binding response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed to get service bindings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service key by ID: %q", serviceKeyID)

	var serviceKey types.ServiceKey
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return nil, errors.Errorf("Failed to unmarshal service key: %v", err)
	}

	return &serviceKey, nil
}

func buildURL(baseURL, path, tenantKey, tenantValue string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Path params
	base.Path += path

	// Query params
	params := url.Values{}
	params.Add(tenantKey, tenantValue)
	base.RawQuery = params.Encode()

	return base.String(), nil
}

func (h *Handler) buildTemplateData(serviceKeyCredentials json.RawMessage, tmReqBody types.TenantMapping) (map[string]string, error) {
	// todo::: consider removing it?
	//svcKeyURL := gjson.Get(string(serviceKeyCredentials), "certificateservice.apiurl").String()
	//if svcKeyURL == "" {
	//	return nil, errors.New("could not find 'certificateservice.apiurl' property")
	//}

	appURL := tmReqBody.Items[0].ApplicationURL

	svcKeyClientID, ok := gjson.Get(string(serviceKeyCredentials), "clientid").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'clientid' property")
	}

	svcKeyClientSecret, ok := gjson.Get(string(serviceKeyCredentials), "clientsecret").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'clientsecret' property")
	}

	svcKeyTokenURL, ok := gjson.Get(string(serviceKeyCredentials), "url").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uaa.url' property")
	}
	//tokenPath := "/oauth/token" // todo::: consider removing it?

	data := map[string]string{
		"URL":          appURL,
		"TokenURL":     svcKeyTokenURL,
		"ClientID":     svcKeyClientID,
		"ClientSecret": svcKeyClientSecret,
		"SubaccountID": h.tenantID,
	}

	return data, nil
}
