package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/api/paths"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/api/types"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/external_caller"
	"github.com/kyma-incubator/compass/components/cim-adapter/pkg/httputil"
	authpkg "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

const (
	SubaccountKey     = "subaccount_id"
	LocationHeaderKey = "Location"
	AssignOperation   = "assign"
	UnassignOperation = "unassign"

	mdiCatalogName = "one-mds"
	mdiPlanName = "sap-integration"
	cimAppNamespace = "sap.cim"
	s4AppNamespace = "s4"
	serviceInstanceMaxLengthName = 50
	serviceBindingMaxLengthName = 100
)

type Handler struct {
	cfg      config.Config
	tenantID string
	caller   *external_caller.Caller
}

func NewHandler(cfg config.Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDForRequest(r)
	errResp := errors.Errorf("An unexpected error occurred while processingg the request. X-Request-Id: %s", correlationID)

	log.C(ctx).Infof("Processing tenantt mapping notification...")
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("Failed to read request body: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Failed to read request body"))
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
	if tm.ReceiverTenant.SubaccountID != "" {
		log.C(ctx).Infof("Use subaccount ID from the request body as tenant")
		h.tenantID = tm.ReceiverTenant.SubaccountID
	} else {
		log.C(ctx).Infof("Use application tennat ID/xsuaa tenant ID from the request body as tenant")
		h.tenantID = tm.ReceiverTenant.ApplicationTenantID
	}

	ic, exists := h.cfg.ServiceManagerCfg.RegionToInstanceConfig[tm.ReceiverTenant.DeploymentRegion]
	if !exists {
		log.C(ctx).Errorf("Missing service manager instance config for region: %s", tm.ReceiverTenant.DeploymentRegion)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}

	securedHTTPClient := authpkg.PrepareHTTPClientWithSSLValidation(h.cfg.HTTPClient.Timeout, h.cfg.HTTPClient.SkipSSLValidation)
	caller, err := external_caller.NewCaller(securedHTTPClient, ic)
	if err != nil {
		log.C(ctx).Errorf("Failed creating external caller: %v", err)
		httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
		return
	}
	h.caller = caller
	h.cfg.ServiceManagerCfg.URL = ic.SMURL

	readyResp := `{"state":"READY"}`
	formationID := tm.Context.FormationID

	if tm.AssignedTenant.ApplicationNamespace != cimAppNamespace && tm.AssignedTenant.ApplicationNamespace != s4AppNamespace {
		err := errors.Errorf("Unexpected assigned tenant application namespace: '%s'. Expected to be either '%s' or '%s'", tm.AssignedTenant.ApplicationNamespace, cimAppNamespace, s4AppNamespace)
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, err)
		return
	}

	if tm.AssignedTenant.ApplicationNamespace == "sap.cim" { // todo::: or maybe use the 'uclSystemName'?
		mdiReadSvcInstanceName := mdiCatalogName + "-read-instance-" + formationID
		if len(mdiReadSvcInstanceName) > serviceInstanceMaxLengthName {
			log.C(ctx).Infof("The length of the service instance name is bigger than %d, truncating it...", serviceInstanceMaxLengthName)
			mdiReadSvcInstanceName = mdiReadSvcInstanceName[:serviceInstanceMaxLengthName]
		}
		if tm.Context.Operation == UnassignOperation {
			log.C(ctx).Infof("Handle MDI 'read' instance for %s operation...", UnassignOperation)
			mdiSvcInstanceID, err := h.retrieveServiceInstanceIDByName(ctx, mdiReadSvcInstanceName)
			if err != nil {
				log.C(ctx).Error(err)
				httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
				return
			}

			if mdiSvcInstanceID != "" {
				if err := h.deleteServiceKeys(ctx, mdiSvcInstanceID, mdiReadSvcInstanceName); err != nil {
					log.C(ctx).Error(err)
					httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
					return
				}
				if err := h.deleteServiceInstance(ctx, mdiSvcInstanceID, mdiReadSvcInstanceName); err != nil {
					log.C(ctx).Error(err)
					httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
					return
				}
			}

			httputil.RespondWithBody(ctx, w, http.StatusOK, json.RawMessage(readyResp))
			return
		}

		log.C(ctx).Infof("Handle MDI 'read' instance for %s operation...", AssignOperation)
		mdiOfferingID, err := h.retrieveServiceOffering(ctx, mdiCatalogName)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiPlanID, err := h.retrieveServicePlan(ctx, mdiPlanName, mdiOfferingID)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiReadInstanceParams := `{"application":"ariba","businessSystemId":"MDCS","enableTenantDeletion":true}`
		svcInstanceIDMDI, err := h.createServiceInstance(ctx, mdiReadSvcInstanceName, mdiPlanID, mdiReadInstanceParams)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiReadSvcKeyName := mdiReadSvcInstanceName + "-key"
		if len(mdiReadSvcKeyName) > serviceBindingMaxLengthName {
			log.C(ctx).Infof("The length of the service binding name is bigger than %d, truncating it...", serviceBindingMaxLengthName)
			mdiReadSvcKeyName = mdiReadSvcKeyName[:serviceBindingMaxLengthName]
		}

		mdiServiceKeyID, err := h.createServiceKey(ctx, mdiReadSvcKeyName, svcInstanceIDMDI, mdiReadSvcInstanceName)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiReadServiceKey, err := h.retrieveServiceKeyByID(ctx, mdiServiceKeyID)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if len(mdiReadServiceKey.Credentials) < 0 {
			log.C(ctx).Errorf("The credentials for service key with ID: %q should not be empty", mdiReadServiceKey.ID)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		data, err := h.buildTemplateData(mdiReadServiceKey.Credentials)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		respBody := `{"state":"READY","configuration":{"credentials":{"outboundCommunication":{"oauth2ClientCredentials":{"url":"{{ .URL }}","tokenServiceUrl":"{{ .TokenURL }}","clientId":"{{ .ClientID }}","clientSecret":"{{ .ClientSecret }}"}}}}}`
		t, err := template.New("").Parse(respBody)
		if err != nil {
			log.C(ctx).Errorf("An error occurred while parsing template: %v", err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		res := new(bytes.Buffer)
		if err = t.Execute(res, data); err != nil {
			log.C(ctx).Errorf("An error occurred while executing template: %v", err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		log.C(ctx).Infof("Successfully processed tenant mapping notification")
		httputil.RespondWithBody(ctx, w, http.StatusOK, json.RawMessage(res.String()))
		return
	}

	if tm.AssignedTenant.ApplicationNamespace == "sap.s4" {
		mdiWriteSvcInstance := mdiCatalogName + "-write-instance-" + formationID
		if len(mdiWriteSvcInstance) > serviceInstanceMaxLengthName {
			log.C(ctx).Infof("The length of the service instance name is bigger than %d, truncating it...", serviceInstanceMaxLengthName)
			mdiWriteSvcInstance = mdiWriteSvcInstance[:serviceInstanceMaxLengthName]
		}

		if tm.Context.Operation == UnassignOperation {
			log.C(ctx).Infof("Handle MDI 'write' instance for %s operation...", UnassignOperation)
			svcInstanceIDMDI, err := h.retrieveServiceInstanceIDByName(ctx, mdiWriteSvcInstance)
			if err != nil {
				log.C(ctx).Error(err)
				httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
				return
			}

			if svcInstanceIDMDI != "" {
				if err := h.deleteServiceKeys(ctx, svcInstanceIDMDI, mdiWriteSvcInstance); err != nil {
					log.C(ctx).Error(err)
					httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
					return
				}
				if err := h.deleteServiceInstance(ctx, svcInstanceIDMDI, mdiWriteSvcInstance); err != nil {
					log.C(ctx).Error(err)
					httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
					return
				}
			}

			httputil.RespondWithBody(ctx, w, http.StatusOK, json.RawMessage(readyResp))
			return
		}

		log.C(ctx).Infof("Handle MDI 'write' instance for %s operation...", AssignOperation)
		offeringIDMDI, err := h.retrieveServiceOffering(ctx, mdiCatalogName)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiPlanID, err := h.retrieveServicePlan(ctx, mdiPlanName, offeringIDMDI)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiWriteInstanceParams := `{"application":"s4","businessSystemId":"MDI","enableTenantDeletion":true,"writePermissions":[{"entityType":"sap.odm.businesspartner.BusinessPartnerRelationship"},{"entityType":"sap.odm.businesspartner.BusinessPartner"},{"entityType":"sap.odm.businesspartner.ContactPersonRelationship"},{"entityType":"sap.odm.finance.costobject.ProjectControllingObject"},{"entityType":"sap.odm.finance.costobject.CostCenter"}]}`
		mdiSvcInstanceID, err := h.createServiceInstance(ctx, mdiWriteSvcInstance, mdiPlanID, mdiWriteInstanceParams)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiWriteSvcKeyName := mdiWriteSvcInstance + "-key"
		if len(mdiWriteSvcKeyName) > serviceBindingMaxLengthName {
			log.C(ctx).Infof("The length of the service binding name is bigger than %d, truncating it...", serviceBindingMaxLengthName)
			mdiWriteSvcKeyName = mdiWriteSvcKeyName[:serviceBindingMaxLengthName]
		}

		mdiWriteServiceKeyID, err := h.createServiceKey(ctx, mdiWriteSvcKeyName, mdiSvcInstanceID, mdiWriteSvcInstance)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		mdiWriteServiceKey, err := h.retrieveServiceKeyByID(ctx, mdiWriteServiceKeyID)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		if len(mdiWriteServiceKey.Credentials) < 0 {
			log.C(ctx).Errorf("The credentials for service key with ID: %q should not be empty", mdiWriteServiceKey.ID)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		data, err := h.buildTemplateData(mdiWriteServiceKey.Credentials)
		if err != nil {
			log.C(ctx).Error(err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		respBody := `{"state":"CONFIG_PENDING","configuration":{"credentials":{"inboundCommunication":{"basicAuthentication":{"correlationIds":["SAP_COM_0594","SAP_COM_0008"],"destinations":[{"name":"mdo-ui","url":"/sap/opu/odata4/sap/mdo_distributionadmin/srvd_a2x/sap/distributionadmin/0001/","additionalProperties":{"MDOProvider":"true","MDOConsumer":"true","MDIInstanceId":"{{ .SystemID }}","MDOBusinessSystem":"{{ .TemplateInput.SourceApplication.Name }}"}},{"name":"{{ .TemplateInput.SourceApplication.Name }}_BPCONFIRM","url":"/sap/bc/srt/scs_ext/sap/businesspartnerrelationshipsu1"}]}},"outboundCommunication":{"basicAuthentication":{"correlationIds":["SAP_COM_0008"],"username":"{{ .ClientID }}","password":"{{ .ClientSecret }}","url":"{{ .URL }}"},"oauth2ClientCredentials":{"correlationIds":["SAP_COM_0659"],"clientId":"{{ .ClientID }}","clientSecret":"{{ .ClientSecret }}","url":"{{ .URL }}","tokenServiceUrl":"{{ .TokenURL }}"}}},"additionalAttributes":{"communicationSystemProperties":[{"name":"Business System","value":"{{ .TemplateInput.SourceApplication.Name }}","correlationIds":["SAP_COM_0659"]},{"name":"Logical System","value":"MDI_BUPA","correlationIds":["SAP_COM_0008"]},{"name":"Business System","value":"{{ .TemplateInput.TargetApplication.Tenant.Labels.subdomain }}","correlationIds":["SAP_COM_0008"]}],"communicationArrangementProperties":[{"name":"Path","value":"/","correlationIds":["SAP_COM_0659"]}],"outboundServicesProperties":[{"name":"Business Partner - Replicate from SAP S/4HANA Cloud to Client","path":"/businesspartner/v0/soap/BusinessPartnerBulkReplicateRequestIn?tenantId={{ .TemplateInput.TargetApplication.Tenant.Labels.subdomain }}","correlationIds":["SAP_COM_0008"],"isServiceActive":true,"additionalProperties":[{"name":"Replication Model","value":"BPMDI_CIM"},{"name":"Replication Mode","value":"C"},{"name":"Output Mode","value":"D"}]},{"name":"Replicate Customers from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Suppliers from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Company Addresses from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Workplace Addresses from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Personal Addresses from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Business Partner Relationship - Replicate from SAP S/4HANA Cloud to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Business Partner - Send Confirmation from SAP S/4HANA Cloud to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"BP Relationship - Send Confirmation from SAP S/4HANA Cloud to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false}]}}}`
		t, err := template.New("").Parse(respBody)
		if err != nil {
			log.C(ctx).Errorf("An error occurred while parsing template: %v", err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		res := new(bytes.Buffer)
		if err = t.Execute(res, data); err != nil {
			log.C(ctx).Errorf("An error occurred while executing template: %v", err)
			httputil.RespondWithError(ctx, w, http.StatusInternalServerError, errResp)
			return
		}

		log.C(ctx).Infof("Successfully processed tenant mapping notification")
		httputil.RespondWithBody(ctx, w, http.StatusOK, json.RawMessage(res.String()))
		return
	}
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func validate(tm types.TenantMapping) error {
	if tm.ReceiverTenant.ApplicationTenantID == "" {
		return errors.New("The subaccount ID in the receiver tenant in the tenant mapping request body should not be empty")
	}

	if tm.ReceiverTenant.DeploymentRegion == "" {
		return errors.New("The deployment region in the receiver tenant in the tenant mapping request body should not be empty")
	}

	if tm.Context.FormationID == "" {
		return errors.New("The formation ID in the tenant mapping request body should not be empty")
	}

	if tm.Context.Operation != AssignOperation && tm.Context.Operation != UnassignOperation {
		return errors.Errorf("The operation in the tenant mapping request body is invalid: %s", tm.Context.Operation)
	}

	return nil
}

func (h *Handler) retrieveServiceOffering(ctx context.Context, catalogName string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceOfferingsPath, SubaccountKey, h.tenantID)
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

	body, err := io.ReadAll(resp.Body)
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

	if offeringID == "" {
		return "", errors.Errorf("Couldn't find service offering for catalog name: %q", catalogName)
	}

	log.C(ctx).Infof("Service offering ID: %q", offeringID)

	return offeringID, nil
}

func (h *Handler) retrieveServicePlan(ctx context.Context, planName, offeringID string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServicePlansPath, SubaccountKey, h.tenantID)
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

	body, err := io.ReadAll(resp.Body)
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

	if planID == "" {
		return "", errors.Errorf("Couldn't find service plan for catalog name: %q and offering ID: %q", planName, offeringID)
	}

	log.C(ctx).Infof("Service plan ID: %q", planID)

	return planID, nil
}

func (h *Handler) createServiceInstance(ctx context.Context, serviceInstanceName, planID, parameters string) (string, error) {
	siReqBody := &types.ServiceInstanceReqBody{
		Name:          serviceInstanceName,
		ServicePlanId: planID,
		Parameters:    json.RawMessage(parameters),
	}

	siReqBodyBytes, err := json.Marshal(siReqBody)
	if err != nil {
		return "", errors.Errorf("Failed to marshal service instance body: %v", err)
	}

	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceInstancesPath, SubaccountKey, h.tenantID)
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

	body, err := io.ReadAll(resp.Body)
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

		opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, h.tenantID)
		if err != nil {
			return "", errors.Wrapf(err, "while building asynchronous service instance operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return "", err
		}

		ticker := time.NewTicker(5 * time.Second)
		timeout := time.After(time.Second * 50)
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service instance with name: %q", serviceInstanceName)
				opResp, err := h.caller.Call(opReq)
				if err != nil {
					return "", err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := io.ReadAll(opResp.Body)
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

func (h *Handler) deleteServiceKeys(ctx context.Context, serviceInstanceID, serviceInstanceName string) error {
	svcKeyIDs, err := h.retrieveServiceKeyIDsByInstanceID(ctx, serviceInstanceID, serviceInstanceName)
	if err != nil {
		return err
	}

	for _, keyID := range svcKeyIDs {
		svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", keyID)
		strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, svcKeyPath, SubaccountKey, h.tenantID)
		if err != nil {
			return errors.Wrapf(err, "while building service binding URL")
		}

		req, err := http.NewRequest(http.MethodDelete, strURL, nil)
		if err != nil {
			return err
		}

		log.C(ctx).Infof("Deleting service binding with ID: %q", keyID)
		resp, err := h.caller.Call(req)
		if err != nil {
			return err
		}
		defer closeResponseBody(ctx, resp)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Errorf("Failed to read response body from service binding deletion request: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			return errors.Errorf("Failed to delete service binding, status: %d, body: %s", resp.StatusCode, body)
		}

		if resp.StatusCode == http.StatusAccepted {
			log.C(ctx).Infof("Handle asynchronous service binding deletion...")
			opStatusPath := resp.Header.Get(LocationHeaderKey)
			if opStatusPath == "" {
				return errors.Errorf("The service binding operation status path from %s header should not be empty", LocationHeaderKey)
			}

			opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, h.tenantID)
			if err != nil {
				return errors.Wrapf(err, "while building asynchronous service binding operation URL")
			}

			opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
			if err != nil {
				return err
			}

			ticker := time.NewTicker(5 * time.Second)
			timeout := time.After(time.Second * 50)
			for {
				select {
				case <-ticker.C:
					log.C(ctx).Infof("Getting asynchronous operation status for service binding with ID: %q", keyID)
					opResp, err := h.caller.Call(opReq)
					if err != nil {
						return err
					}
					defer closeResponseBody(ctx, opResp)

					opBody, err := io.ReadAll(opResp.Body)
					if err != nil {
						return errors.Errorf("Failed to read operation response body from asynchronous service binding deletion request: %v", err)
					}

					if opResp.StatusCode != http.StatusOK {
						return errors.Errorf("Failed to get asynchronous service binding operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
					}

					var opStatus types.OperationStatus
					err = json.Unmarshal(opBody, &opStatus)
					if err != nil {
						return errors.Errorf("Failed to unmarshal service binding operation status: %v", err)
					}

					if opStatus.State == types.OperationStateInProgress {
						log.C(ctx).Infof("The asynchronous service binding operation state is still: %q", types.OperationStateInProgress)
						continue
					}

					if opStatus.State != types.OperationStateSucceeded {
						return errors.Errorf("The asynchronous service binding operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
					}

					log.C(ctx).Infof("The asynchronous operation status for service binding with ID: %q finished with state: %q", keyID, opStatus.State)
					return nil
				case <-timeout:
					return errors.New("Timeout waiting for asynchronous operation status to finish")
				}
			}
		}

		log.C(ctx).Infof("Successfully deleted service binding with ID: %q synchronously", keyID)
	}

	return nil
}

func (h *Handler) deleteServiceInstance(ctx context.Context, serviceInstanceID, serviceInstanceName string) error {
	svcInstancePath := paths.ServiceInstancesPath + fmt.Sprintf("/%s", serviceInstanceID)
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, svcInstancePath, SubaccountKey, h.tenantID)
	if err != nil {
		return errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodDelete, strURL, nil)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Deleting service instance with ID: %q and name: %q", serviceInstanceID, serviceInstanceName)
	resp, err := h.caller.Call(req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("Failed to read response body from service instance deletion request: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errors.Errorf("Failed to delete service instance, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous service instance deletion...")
		opStatusPath := resp.Header.Get(LocationHeaderKey)
		if opStatusPath == "" {
			return errors.Errorf("The service instance operation status path from %s header should not be empty", LocationHeaderKey)
		}

		opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, h.tenantID)
		if err != nil {
			return errors.Wrapf(err, "while building asynchronous service instance operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return err
		}

		ticker := time.NewTicker(5 * time.Second)
		timeout := time.After(time.Second * 50)
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service instance with ID: %q and name: %q", serviceInstanceID, serviceInstanceName)
				opResp, err := h.caller.Call(opReq)
				if err != nil {
					return err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := io.ReadAll(opResp.Body)
				if err != nil {
					return errors.Errorf("Failed to read operation response body from asynchronous service instance deletion request: %v", err)
				}

				if opResp.StatusCode != http.StatusOK {
					return errors.Errorf("Failed to get asynchronous service instance operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
				}

				var opStatus types.OperationStatus
				err = json.Unmarshal(opBody, &opStatus)
				if err != nil {
					return errors.Errorf("Failed to unmarshal service instance operation status: %v", err)
				}

				if opStatus.State == types.OperationStateInProgress {
					log.C(ctx).Infof("The asynchronous service instance operation state is still: %q", types.OperationStateInProgress)
					continue
				}

				if opStatus.State != types.OperationStateSucceeded {
					return errors.Errorf("The asynchronous service instance operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
				}

				log.C(ctx).Infof("The asynchronous operation status for service instance with name: %q finished with state: %q", serviceInstanceName, opStatus.State)
				return nil
			case <-timeout:
				return errors.New("Timeout waiting for asynchronous operation status to finish")
			}
		}
	}

	log.C(ctx).Infof("Successfully deleted service instance with ID: %q synchronously", serviceInstanceID)

	return nil
}

func (h *Handler) retrieveServiceInstanceIDByName(ctx context.Context, serviceInstanceName string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceInstancesPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Info("Listing service instances...")
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("Failed to read service instances response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Failed to get service instances, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service instances")

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

	if instanceID == "" {
		log.C(ctx).Warnf("No instance ID found by name: %q", serviceInstanceName)
		return "", nil
	}

	log.C(ctx).Infof("Successfully find service instance ID: %q by instance name: %q", instanceID, serviceInstanceName)
	return instanceID, nil
}

func (h *Handler) createServiceKey(ctx context.Context, serviceKeyName, serviceInstanceID, serviceInstanceNameProcurement string) (string, error) {
	serviceKeyReqBody := &types.ServiceKeyReqBody{
		Name:              serviceKeyName,
		ServiceInstanceId: serviceInstanceID,
	}

	serviceKeyReqBodyBytes, err := json.Marshal(serviceKeyReqBody)
	if err != nil {
		return "", errors.Errorf("Failed to marshal service key body: %v", err)
	}

	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceBindingsPath, SubaccountKey, h.tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service bindings URL")
	}

	log.C(ctx).Infof("Creating service key for service instance with name: %q", serviceInstanceNameProcurement)
	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(serviceKeyReqBodyBytes))
	if err != nil {
		return "", err
	}

	resp, err := h.caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
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

		opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, h.tenantID)
		if err != nil {
			return "", errors.Wrapf(err, "while building asynchronous service key operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return "", err
		}

		ticker := time.NewTicker(5 * time.Second)
		timeout := time.After(time.Second * 50)
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service key with name: %q", serviceKeyName)
				opResp, err := h.caller.Call(opReq)
				if err != nil {
					return "", err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := io.ReadAll(opResp.Body)
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

func (h *Handler) retrieveServiceKeyIDsByInstanceID(ctx context.Context, serviceInstanceID, serviceInstanceName string) ([]string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceBindingsPath, SubaccountKey, h.tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building service binding URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Listing service bindings for instance with ID: %q and name: %q", serviceInstanceID, serviceInstanceName)
	resp, err := h.caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("Failed to read service binding response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed to get service bindings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service bindings for instance with ID: %q and name: %q", serviceInstanceID, serviceInstanceName)

	var svcKeys types.ServiceKeys
	err = json.Unmarshal(body, &svcKeys)
	if err != nil {
		return nil, errors.Errorf("Failed to unmarshal service keys: %v", err)
	}

	serviceKeysIDs := make([]string, 0, len(svcKeys.Items))
	for _, key := range svcKeys.Items {
		if key.ServiceInstanceId == serviceInstanceID {
			serviceKeysIDs = append(serviceKeysIDs, key.ID)
		}
	}
	log.C(ctx).Infof("Service instance with ID: %q and name: %q has/have %d keys(s)", serviceInstanceID, serviceInstanceName, len(serviceKeysIDs))

	return serviceKeysIDs, nil
}

func (h *Handler) retrieveServiceKeyByID(ctx context.Context, serviceKeyID string) (*types.ServiceKey, error) {
	svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", serviceKeyID)
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, svcKeyPath, SubaccountKey, h.tenantID)
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

	body, err := io.ReadAll(resp.Body)
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

func (h *Handler) buildTemplateData(serviceKeyCredentials json.RawMessage) (map[string]string, error) {
	systemID, ok := gjson.Get(string(serviceKeyCredentials), "systemId").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'systemId' property")
	}

	svcKeyURI, ok := gjson.Get(string(serviceKeyCredentials), "uri").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uri' property")
	}

	svcKeyClientID, ok := gjson.Get(string(serviceKeyCredentials), "uaa.clientid").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uaa.clientid' property")
	}

	svcKeyClientSecret, ok := gjson.Get(string(serviceKeyCredentials), "uaa.clientsecret").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uaa.clientsecret' property")
	}

	svcKeyTokenURL, ok := gjson.Get(string(serviceKeyCredentials), "uaa.url").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uaa.url' property")
	}
	//tokenPath := "/oauth/token" // todo::: check do we need the suffix in the destination

	data := map[string]string{
		"SystemID":     systemID,
		"URL":          svcKeyURI,
		"TokenURL":     svcKeyTokenURL,
		"ClientID":     svcKeyClientID,
		"ClientSecret": svcKeyClientSecret,
	}

	return data, nil
}

func isConfigNonEmpty(configuration string) bool {
	if configuration != "" && configuration != "{}" && configuration != "\"\"" && configuration != "null" {
		return true
	}

	return false
}
