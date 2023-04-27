package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/paths"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/types"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/external_caller"
	"github.com/kyma-incubator/compass/components/tm-adapter/pkg/httputil"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type Handler struct {
	cfg    *config.Config
	caller *external_caller.Caller
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
	log.C(ctx).Infof("Tenant mapping request body: %q", reqBody)

	//var tm types.TenantMapping
	//err = json.Unmarshal(reqBody, &tm)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to unmarshal request body: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New( "Invalid json"))
	//	return
	//}

	catalogName := "certificate-service" // todo::: should be provided as label on the runtime and will be used through TM notification body
	planName := "standard"               // todo::: should be provided as label on the runtime and will be used through TM notification body

	offeringID, err := h.retrieveServiceOffering(ctx, catalogName)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	planID, err := h.retrieveServicePlan(ctx, planName, offeringID)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	serviceInstanceName := catalogName + "-instance"
	serviceInstanceID, err := h.createServiceInstance(ctx, planID, serviceInstanceName)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	serviceKeyName := serviceInstanceName + "-key"
	serviceKey, err := h.createServiceKey(ctx, serviceKeyName, serviceInstanceID)
	if err != nil {
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	log.C(ctx).Infof("Service key creds: %v", string(serviceKey.Credentials)) // todo::: remove

	//// get service offerings
	//log.C(ctx).Infof("Listing service offerings...")
	//req, err := http.NewRequest(http.MethodGet, h.cfg.ServiceManagerURL+paths.ServiceOfferingsPath, nil)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//resp, err := h.caller.Call(req)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//defer closeResponseBody(ctx, resp)
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to read service offerings response body: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//if resp.StatusCode != http.StatusOK {
	//	errMsg := fmt.Sprintf("Failed to get service offerings, status: %d, body: %q", resp.StatusCode, body)
	//	log.C(ctx).Error(errMsg)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New(errMsg))
	//	return
	//}
	//log.C(ctx).Infof("offerings resp --> %s", string(body)) // todo::: remove
	//log.C(ctx).Infof("Successfully fetch service offerings")
	//
	//var offerings types.ServiceOfferings
	//err = json.Unmarshal(body, &offerings)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to unmarshal service offerings: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//var offeringID string
	//for _, item := range offerings.Items {
	//	if item.CatalogName == catalogName {
	//		offeringID = item.Id
	//		break
	//	}
	//}
	//log.C(ctx).Infof("Service offering ID: %q", offeringID)
	//
	//// get service plans
	//log.C(ctx).Infof("Listing service plans...")
	//req, err = http.NewRequest(http.MethodGet, h.cfg.ServiceManagerURL+paths.ServicePlansPath, nil)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//resp, err = h.caller.Call(req)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//defer closeResponseBody(ctx, resp)
	//
	//body, err = ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to read service plans response body: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//if resp.StatusCode != http.StatusOK {
	//	errMsg := fmt.Sprintf("Failed to get service plans, status: %d, body: %q", resp.StatusCode, body)
	//	log.C(ctx).Error(errMsg)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New(errMsg))
	//	return
	//}
	//log.C(ctx).Infof("plans resp --> %s", string(body)) // todo::: remove
	//log.C(ctx).Infof("Successfully fetch service plans")
	//
	//var plans types.ServicePlans
	//err = json.Unmarshal(body, &plans)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to unmarshal service plans: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//var planID string
	//for _, item := range plans.Items {
	//	if item.CatalogName == planName && item.ServiceOfferingId == offeringID {
	//		planID = item.Id
	//		break
	//	}
	//}
	//log.C(ctx).Infof("Service plan ID: %q", planID)
	//
	//// create service instance
	//serviceInstanceName := "test-instance-name-api"
	//siReqBody := &types.ServiceInstanceReqBody{
	//	Name:          serviceInstanceName,
	//	ServicePlanId: planID,
	//}
	//
	//siReqBodyBytes, err := json.Marshal(siReqBody)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to marshal service instance body: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//log.C(ctx).Infof("Creating service instance with name: %q", serviceInstanceName)
	//req, err = http.NewRequest(http.MethodPost, h.cfg.ServiceManagerURL+paths.ServiceInstancesPath, bytes.NewBuffer(siReqBodyBytes))
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//resp, err = h.caller.Call(req)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//defer closeResponseBody(ctx, resp)
	//
	//body, err = ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to read response body from service instance creation request: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//if resp.StatusCode != http.StatusCreated {
	//	errMsg := fmt.Sprintf("Failed to create service instance, status: %d, body: %q", resp.StatusCode, body)
	//	log.C(ctx).Error(errMsg)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New(errMsg))
	//	return
	//}
	//log.C(ctx).Infof("create service instance resp --> %s", string(body)) // todo::: remove
	//log.C(ctx).Infof("Successfully create service instance with name: %q", serviceInstanceName)
	//
	//var serviceInstance types.ServiceInstance
	//err = json.Unmarshal(body, &serviceInstance)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to unmarshal service instance: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//serviceInstanceID := serviceInstance.Id
	//log.C(ctx).Infof("Service instance ID: %q", serviceInstanceID)
	//
	//// create service binding
	//serviceKeyName := serviceInstanceName+"-key"
	//serviceKeyReqBody := &types.ServiceKeyReqBody{
	//	Name:              serviceKeyName,
	//	ServiceInstanceId: serviceInstanceID,
	//	//Parameters: // todo::: should be provided as `parameters` label in the TM notification body - `receiverTenant.parameters`?
	//}
	//
	//serviceKeyReqBodyBytes, err := json.Marshal(serviceKeyReqBody)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to marshal service key body: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//log.C(ctx).Infof("Creating service key with name: %q from service instance with name: %q and ID: %q", serviceKeyName, serviceInstanceName, serviceInstanceID)
	//req, err = http.NewRequest(http.MethodPost, h.cfg.ServiceManagerURL+paths.ServiceBindingsPath, bytes.NewBuffer(serviceKeyReqBodyBytes))
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//resp, err = h.caller.Call(req)
	//if err != nil {
	//	log.C(ctx).Error(err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//defer closeResponseBody(ctx, resp)
	//
	//body, err = ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to read response body from service key creation request: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//if resp.StatusCode != http.StatusCreated {
	//	errMsg := fmt.Sprintf("Failed to create service key, status: %d, body: %q", resp.StatusCode, body)
	//	log.C(ctx).Error(errMsg)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New(errMsg))
	//	return
	//}
	//log.C(ctx).Infof("create service key resp --> %s", string(body)) // todo::: remove
	//log.C(ctx).Infof("Successfully create service key with name: %q", serviceKeyName)
	//
	//var serviceKey types.ServiceKey
	//err = json.Unmarshal(body, &serviceKey)
	//if err != nil {
	//	log.C(ctx).Errorf("Failed to unmarshal service key: %v", err)
	//	httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
	//	return
	//}
	//
	//serviceKeyCreds := serviceKey.Credentials
	//log.C(ctx).Infof("Service key ID: %q", serviceKey.Id)
	//log.C(ctx).Infof("Service key creds: %v", string(serviceKeyCreds)) // todo::: remove

	log.C(ctx).Infof("Successfully processed tenant mapping notification")
	httputil.Respond(w, http.StatusOK)
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func (h *Handler) retrieveServiceOffering(ctx context.Context, catalogName string) (string, error) {
	log.C(ctx).Infof("Listing service offerings...")
	req, err := http.NewRequest(http.MethodGet, h.cfg.ServiceManagerURL+paths.ServiceOfferingsPath, nil)
	if err != nil {
		return "", err
	}

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
		return "", errors.Errorf("Failed to get service offerings, status: %d, body: %q", resp.StatusCode, body)
	}
	log.C(ctx).Infof("offerings resp --> %s", string(body)) // todo::: remove
	log.C(ctx).Infof("Successfully fetch service offerings")

	var offerings types.ServiceOfferings
	err = json.Unmarshal(body, &offerings)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service offerings: %v", err)
	}

	var offeringID string
	for _, item := range offerings.Items {
		if item.CatalogName == catalogName {
			offeringID = item.Id
			break
		}
	}
	log.C(ctx).Infof("Service offering ID: %q", offeringID)

	return offeringID, nil
}

func (h *Handler) retrieveServicePlan(ctx context.Context, planName, offeringID string) (string, error) {
	log.C(ctx).Infof("Listing service plans...")
	req, err := http.NewRequest(http.MethodGet, h.cfg.ServiceManagerURL+paths.ServicePlansPath, nil)
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
		return "", errors.Errorf("Failed to read service plans response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Failed to get service plans, status: %d, body: %q", resp.StatusCode, body)
	}
	log.C(ctx).Infof("plans resp --> %s", string(body)) // todo::: remove
	log.C(ctx).Infof("Successfully fetch service plans")

	var plans types.ServicePlans
	err = json.Unmarshal(body, &plans)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service plans: %v", err)
	}

	var planID string
	for _, item := range plans.Items {
		if item.CatalogName == planName && item.ServiceOfferingId == offeringID {
			planID = item.Id
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

	log.C(ctx).Infof("Creating service instance with name: %q from plan with ID: %q", serviceInstanceName, planID)
	req, err := http.NewRequest(http.MethodPost, h.cfg.ServiceManagerURL+paths.ServiceInstancesPath, bytes.NewBuffer(siReqBodyBytes))
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
		return "", errors.Errorf("Failed to read response body from service instance creation request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.Errorf("Failed to create service instance, status: %d, body: %q", resp.StatusCode, body)
	}
	log.C(ctx).Infof("create service instance resp --> %s", string(body)) // todo::: remove
	log.C(ctx).Infof("Successfully create service instance with name: %q", serviceInstanceName)

	var serviceInstance types.ServiceInstance
	err = json.Unmarshal(body, &serviceInstance)
	if err != nil {
		return "", errors.Errorf("Failed to unmarshal service instance: %v", err)
	}

	serviceInstanceID := serviceInstance.Id
	log.C(ctx).Infof("Service instance ID: %q", serviceInstanceID)

	return serviceInstanceID, nil
}

func (h *Handler) createServiceKey(ctx context.Context, serviceKeyName, serviceInstanceID string) (*types.ServiceKey, error) {
	serviceKeyReqBody := &types.ServiceKeyReqBody{
		Name:              serviceKeyName,
		ServiceInstanceId: serviceInstanceID,
		//Parameters: // todo::: should be provided as `parameters` label in the TM notification body - `receiverTenant.parameters`?
	}

	serviceKeyReqBodyBytes, err := json.Marshal(serviceKeyReqBody)
	if err != nil {
		return nil, errors.Errorf("Failed to marshal service key body: %v", err)
	}

	log.C(ctx).Infof("Creating service key with name: %q from service instance with ID: %q", serviceKeyName, serviceInstanceID)
	req, err := http.NewRequest(http.MethodPost, h.cfg.ServiceManagerURL+paths.ServiceBindingsPath, bytes.NewBuffer(serviceKeyReqBodyBytes))
	if err != nil {
		return nil, err
	}

	resp, err := h.caller.Call(req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("Failed to read response body from service key creation request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("Failed to create service key, status: %d, body: %q", resp.StatusCode, body)
	}
	log.C(ctx).Infof("create service key resp --> %s", string(body)) // todo::: remove
	log.C(ctx).Infof("Successfully create service key with name: %q", serviceKeyName)

	var serviceKey types.ServiceKey
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return nil, errors.Errorf("Failed to unmarshal service key: %v", err)
	}

	log.C(ctx).Infof("Service key ID: %q", serviceKey.Id)

	return &serviceKey, nil
}
