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
	"net/url"
)

const subaccountKey = "subaccount_id"

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
	log.C(ctx).Infof("Tenant mapping request body: %q", reqBody)

	var tm types.TenantMapping
	err = json.Unmarshal(reqBody, &tm)
	if err != nil {
		log.C(ctx).Errorf("Failed to unmarshal request body: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Invalid json"))
		return
	}

	catalogName := "certificate-service" // todo::: should be provided as label on the runtime/app-template and will be used through TM notification body
	planName := "standard"               // todo::: should be provided as label on the runtime/app-template and will be used through TM notification body
	h.tenantID = tm.ReceiverTenant.SubaccountID

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

	responseBody := types.Response{Configuration: serviceKey.Credentials}

	log.C(ctx).Infof("Successfully processed tenant mapping notification")
	httputil.RespondWithBody(ctx, w, http.StatusOK, responseBody)
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func (h *Handler) retrieveServiceOffering(ctx context.Context, catalogName string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceOfferingsPath, subaccountKey, h.tenantID)
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
		return "", errors.Errorf("Failed to get service offerings, status: %d, body: %q", resp.StatusCode, body)
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
			offeringID = item.Id
			break
		}
	}
	log.C(ctx).Infof("Service offering ID: %q", offeringID)

	return offeringID, nil
}

func (h *Handler) retrieveServicePlan(ctx context.Context, planName, offeringID string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServicePlansPath, subaccountKey, h.tenantID)
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
		return "", errors.Errorf("Failed to get service plans, status: %d, body: %q", resp.StatusCode, body)
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

	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceInstancesPath, subaccountKey, h.tenantID)
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

	if resp.StatusCode != http.StatusCreated {
		return "", errors.Errorf("Failed to create service instance, status: %d, body: %q", resp.StatusCode, body)
	}
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

	strURL, err := buildURL(h.cfg.ServiceManagerURL, paths.ServiceBindingsPath, subaccountKey, h.tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building service bindings URL")
	}

	log.C(ctx).Infof("Creating service key with name: %q from service instance with ID: %q", serviceKeyName, serviceInstanceID)
	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(serviceKeyReqBodyBytes))
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
	log.C(ctx).Infof("Successfully create service key with name: %q", serviceKeyName)

	var serviceKey types.ServiceKey
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return nil, errors.Errorf("Failed to unmarshal service key: %v", err)
	}

	log.C(ctx).Infof("Service key ID: %q", serviceKey.Id)

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
