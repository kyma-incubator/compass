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
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	SubaccountKey              = "subaccount_id"
	LocationHeaderKey          = "Location"
	AssignOperation            = "assign"
	UnassignOperation          = "unassign"
	contentTypeHeaderKey       = "unassign"
	contentTypeApplicationJSON = "unassign"

	mdiCatalogName     = "one-mds"
	mdiPlanName        = "sap-integration"
	billingCatalogName = "SAPHybrisRevenueCloud"
	billingPlanName    = "default"

	cimAppNamespace     = "sap.cim"
	mdoAppNamespace     = "sap.mdo"
	s4AppNamespace      = "sap.s4"
	billingAppNamespace = "sap.billing"

	serviceInstanceNameMaxLength = 50
	serviceBindingMaxLengthName  = 100

	beginCertTag = "-----BEGIN CERTIFICATE-----"
	endCertTag   = "-----END CERTIFICATE-----"
)

type Handler struct {
	cfg            config.Config
	mtlsHTTPClient *http.Client
	mu             *sync.RWMutex
	caller         *external_caller.Caller
}

func NewHandler(cfg config.Config, mtlsHTTPClient *http.Client) *Handler {
	return &Handler{
		cfg:            cfg,
		mtlsHTTPClient: mtlsHTTPClient,
		mu:             &sync.RWMutex{},
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
	var tenantID string
	if tm.ReceiverTenant.SubaccountID != "" {
		log.C(ctx).Infof("Use subaccount ID from the request body as tenant")
		tenantID = tm.ReceiverTenant.SubaccountID
	} else {
		log.C(ctx).Infof("Use application tennat ID/xsuaa tenant ID from the request body as tenant")
		tenantID = tm.ReceiverTenant.ApplicationTenantID
	}

	log.C(ctx).Infof(tm.Print())

	statusAPIURL := r.Header.Get(LocationHeaderKey)
	if statusAPIURL == "" {
		err := errors.Errorf("The value of the %s header could not be empty", LocationHeaderKey)
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, err)
		return
	}

	// This is due to MDO configuration and because it has references to the reverse service instance config.
	// It that case we should return only CONFIG_PENDING state
	if tm.AssignedTenant.ApplicationNamespace == mdoAppNamespace && tm.ReceiverTenant.ApplicationNamespace == s4AppNamespace {
		log.C(ctx).Infof("Service instances details are not provided from the mdo side but from S4. Returning CONFIG_PENDING state")
		httputil.Respond(w, http.StatusAccepted)
		go func() {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			correlationIDKey := correlation.RequestIDHeaderKey
			ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

			logger := log.C(ctx).WithField(correlationIDKey, correlationID)
			ctx = log.ContextWithLogger(ctx, logger)

			reqBody := "{\"state\":\"CONFIG_PENDING\"}"
			if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
				log.C(ctx).Error(statusAPIErr)
			}
		}()
		return
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

	readyResp := "{\"state\":\"READY\"}"
	formationID := tm.Context.FormationID

	if tm.AssignedTenant.ApplicationNamespace != cimAppNamespace && tm.AssignedTenant.ApplicationNamespace != s4AppNamespace {
		err := errors.Errorf("Unexpected assigned tenant application namespace: '%s'. Expected to be either '%s' or '%s'", tm.AssignedTenant.ApplicationNamespace, cimAppNamespace, s4AppNamespace)
		log.C(ctx).Error(err)
		httputil.RespondWithError(ctx, w, http.StatusBadRequest, err)
		return
	}

	httputil.Respond(w, http.StatusAccepted)

	go func(m *sync.RWMutex) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		correlationIDKey := correlation.RequestIDHeaderKey
		ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

		logger := log.C(ctx).WithField(correlationIDKey, correlationID)
		ctx = log.ContextWithLogger(ctx, logger)

		if tm.AssignedTenant.ApplicationNamespace == cimAppNamespace && tm.ReceiverTenant.ApplicationNamespace == mdoAppNamespace {
			mdiReadSvcInstanceName := mdiCatalogName + "-read-instance-" + formationID
			mdiReadSvcInstanceName = truncateString(ctx, mdiReadSvcInstanceName, serviceInstanceNameMaxLength)

			if tm.Context.Operation == UnassignOperation {
				log.C(ctx).Infof("Handle MDI 'read' instance for %s operation...", UnassignOperation)
				mdiReadSvcInstanceID, err := h.retrieveServiceInstanceIDByName(ctx, mdiReadSvcInstanceName, tenantID)
				if err != nil {
					log.C(ctx).Error(err)
					reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
					if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
						log.C(ctx).Error(statusAPIErr)
					}
					return
				}

				if mdiReadSvcInstanceID != "" {
					if err := h.deleteServiceKeys(ctx, mdiReadSvcInstanceID, mdiReadSvcInstanceName, tenantID); err != nil {
						log.C(ctx).Error(err)
						reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while deleting service key(s) for MDI 'read' instance: %s\"}", err.Error())
						if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
							log.C(ctx).Error(statusAPIErr)
						}
						return
					}
					if err := h.deleteServiceInstance(ctx, mdiReadSvcInstanceID, mdiReadSvcInstanceName, tenantID); err != nil {
						log.C(ctx).Error(err)
						reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while deleting MDI 'read' instance: %s\"}", err.Error())
						if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
							log.C(ctx).Error(statusAPIErr)
						}
						return
					}
				}

				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, readyResp); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			log.C(ctx).Infof("Handle MDI 'read' instance for %s operation...", AssignOperation)
			mdiOfferingID, err := h.retrieveServiceOffering(ctx, mdiCatalogName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			mdiPlanID, err := h.retrieveServicePlan(ctx, mdiPlanName, mdiOfferingID, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving service plans: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			m.Lock()
			mdiReadSvcInstanceID, err := h.retrieveServiceInstanceIDByName(ctx, mdiReadSvcInstanceName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			if mdiReadSvcInstanceID != "" {
				log.C(ctx).Infof("Service instance with name: %s is already created or in process of creating. Returning...", mdiReadSvcInstanceName)
				return
			}

			mdiReadInstanceParams := `{"application":"ariba","businessSystemId":"MDCS","enableTenantDeletion":true}`
			mdiReadSvcInstanceID, err = h.createServiceInstance(ctx, mdiReadSvcInstanceName, mdiPlanID, mdiReadInstanceParams, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while creating MDI 'read' service instance: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			mdiReadSvcKeyName := mdiReadSvcInstanceName + "-key"
			mdiReadSvcKeyName = truncateString(ctx, mdiReadSvcKeyName, serviceBindingMaxLengthName)

			mdiServiceKeyID, err := h.createServiceKey(ctx, mdiReadSvcKeyName, "", mdiReadSvcInstanceID, mdiReadSvcInstanceName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while creating MDI 'read' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}
			m.Unlock()

			mdiReadServiceKey, err := h.retrieveServiceKeyByID(ctx, mdiServiceKeyID, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving MDI 'read' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			if len(mdiReadServiceKey.Credentials) < 0 {
				log.C(ctx).Errorf("The credentials for MDI 'read' service key with ID: %q should not be empty", mdiReadServiceKey.ID)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"The service key for the MDI 'read' instance shoud not be empty: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			data, err := h.buildTemplateData(mdiReadServiceKey.Credentials)
			if err != nil {
				log.C(ctx).Errorf("An error occurred while building template data with the MDI 'read' service key: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while building template data with the MDI 'read' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			respBody := `{"state":"READY","configuration":{"credentials":{"outboundCommunication":{"oauth2ClientCredentials":{"url":"{{ .URL }}","tokenServiceUrl":"{{ .TokenURL }}","clientId":"{{ .ClientID }}","clientSecret":"{{ .ClientSecret }}"}}}}}`
			t, err := template.New("").Parse(respBody)
			if err != nil {
				log.C(ctx).Errorf("An error occurred while parsing template: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while parsing MDI 'read' template: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			res := new(bytes.Buffer)
			if err = t.Execute(res, data); err != nil {
				log.C(ctx).Errorf("An error occurred while executing template: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while executing MDI 'read' template: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			log.C(ctx).Infof("Successfully processed tenant mapping notification about CIM")
			if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, res.String()); statusAPIErr != nil {
				log.C(ctx).Error(statusAPIErr)
			}
			return
		}

		if tm.AssignedTenant.ApplicationNamespace == s4AppNamespace && tm.ReceiverTenant.ApplicationNamespace == mdoAppNamespace {
			mdiWriteSvcInstanceName := mdiCatalogName + "-write-instance-" + formationID
			mdiWriteSvcInstanceName = truncateString(ctx, mdiWriteSvcInstanceName, serviceInstanceNameMaxLength)

			if tm.Context.Operation == UnassignOperation {
				log.C(ctx).Infof("Handle MDI 'write' instance for %s operation...", UnassignOperation)
				mdiWriteSvcInstanceID, err := h.retrieveServiceInstanceIDByName(ctx, mdiWriteSvcInstanceName, tenantID)
				if err != nil {
					log.C(ctx).Error(err)
					reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
					if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
						log.C(ctx).Error(statusAPIErr)
					}
					return
				}

				if mdiWriteSvcInstanceID != "" {
					if err := h.deleteServiceKeys(ctx, mdiWriteSvcInstanceID, mdiWriteSvcInstanceName, tenantID); err != nil {
						log.C(ctx).Error(err)
						reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while deleting service key(s) for MDI 'write' instance: %s\"}", err.Error())
						if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
							log.C(ctx).Error(statusAPIErr)
						}
						return
					}
					if err := h.deleteServiceInstance(ctx, mdiWriteSvcInstanceID, mdiWriteSvcInstanceName, tenantID); err != nil {
						log.C(ctx).Error(err)
						reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while deleting MDI 'write' instance: %s\"}", err.Error())
						if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
							log.C(ctx).Error(statusAPIErr)
						}
						return
					}
				}

				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, readyResp); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			log.C(ctx).Infof("Handle MDI 'write' instance for %s operation...", AssignOperation)

			var businessSystemID string
			isAssignedTntCfgNonEmpty := isConfigNonEmpty(string(tm.AssignedTenant.Configuration))
			if tm.Context.Operation == AssignOperation && tm.AssignedTenant.State == "CONFIG_PENDING" && tm.AssignedTenant.Configuration != nil && isAssignedTntCfgNonEmpty {
				log.C(ctx).Infof("Notification request is received for %q operation with CONFIG_PENDING state of the assigned tenant and service instances details. Checking for businessSystemId...", AssignOperation)
				bizSystemID := gjson.GetBytes(tm.AssignedTenant.Configuration, "credentials.inboundCommunication.serviceInstances.0.parameters.businessSystemId").String()
				if bizSystemID == "" {
					log.C(ctx).Error("The business system ID property in the assigned tenant configuration cannot be empty")
					reqBody := "{\"state\":\"CREATE_ERROR\", \"error\": \"The business system ID property in the assigned tenant configuration cannot be empty\"}"
					if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
						log.C(ctx).Error(statusAPIErr)
					}
					return
				}
				businessSystemID = bizSystemID
			}

			if businessSystemID == "" {
				log.C(ctx).Error("The business system ID cannot be empty")
				return
			}

			if tm.Context.Operation == AssignOperation && tm.ReceiverTenant.Subdomain == "" {
				log.C(ctx).Error("The receiver subdomain cannot be empty")
				reqBody := "{\"state\":\"CREATE_ERROR\", \"error\": \"The receiver subdomain cannot be empty\"}"
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			offeringIDMDI, err := h.retrieveServiceOffering(ctx, mdiCatalogName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving service offerings: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			mdiPlanID, err := h.retrieveServicePlan(ctx, mdiPlanName, offeringIDMDI, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving service plans: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			m.Lock()
			mdiWriteSvcInstanceID, err := h.retrieveServiceInstanceIDByName(ctx, mdiWriteSvcInstanceName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			if mdiWriteSvcInstanceID != "" {
				log.C(ctx).Infof("Service instance with name: %s is already created or in process of creating. Returning...", mdiWriteSvcInstanceName)
				return
			}

			mdiWriteInstanceParams := fmt.Sprintf(`{"application":"s4","businessSystemId":"%s","enableTenantDeletion":true,"writePermissions":[{"entityType":"sap.odm.businesspartner.BusinessPartnerRelationship"},{"entityType":"sap.odm.businesspartner.BusinessPartner"},{"entityType":"sap.odm.businesspartner.ContactPersonRelationship"},{"entityType":"sap.odm.finance.costobject.ProjectControllingObject"},{"entityType":"sap.odm.finance.costobject.CostCenter"}]}`, businessSystemID)
			mdiWriteSvcInstanceID, err = h.createServiceInstance(ctx, mdiWriteSvcInstanceName, mdiPlanID, mdiWriteInstanceParams, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while creating MDI 'write' service instance: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			mdiWriteSvcKeyName := mdiWriteSvcInstanceName + "-key"
			mdiWriteSvcKeyName = truncateString(ctx, mdiWriteSvcKeyName, serviceBindingMaxLengthName)

			mdiWriteServiceKeyID, err := h.createServiceKey(ctx, mdiWriteSvcKeyName, "", mdiWriteSvcInstanceID, mdiWriteSvcInstanceName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while creating MDI 'write' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}
			m.Unlock()

			mdiWriteServiceKey, err := h.retrieveServiceKeyByID(ctx, mdiWriteServiceKeyID, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving MDI 'write' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			if len(mdiWriteServiceKey.Credentials) < 0 {
				log.C(ctx).Errorf("The credentials for MDI 'write' service key with ID: %q should not be empty", mdiWriteServiceKey.ID)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"The service key for the MDI 'write' instance shoud not be empty: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			data, err := h.buildTemplateData(mdiWriteServiceKey.Credentials)
			if err != nil {
				log.C(ctx).Errorf("An error occurred while building template data with the MDI 'write' service key: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while building template data with the MDI 'write' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			receiverSubdomain := tm.ReceiverTenant.Subdomain
			respBody := fmt.Sprintf(`{"state":"CONFIG_PENDING","configuration":{"credentials":{"inboundCommunication":{"basicAuthentication":{"correlationIds":["SAP_COM_0594","SAP_COM_0008"],"destinations":[{"name":"mdo-ui","url":"/sap/opu/odata4/sap/mdo_distributionadmin/srvd_a2x/sap/distributionadmin/0001/","additionalProperties":{"MDOProvider":"true","MDOConsumer":"true","MDIInstanceId":"{{ .SystemID }}","MDOBusinessSystem":"%s"}},{"name":"%s_BPCONFIRM","url":"/sap/bc/srt/scs_ext/sap/businesspartnerrelationshipsu1"}]}},"outboundCommunication":{"basicAuthentication":{"correlationIds":["SAP_COM_0008"],"username":"{{ .ClientID }}","password":"{{ .ClientSecret }}","url":"{{ .URL }}"},"oauth2ClientCredentials":{"correlationIds":["SAP_COM_0659"],"clientId":"{{ .ClientID }}","clientSecret":"{{ .ClientSecret }}","url":"{{ .URL }}","tokenServiceUrl":"{{ .TokenURL }}"}}},"additionalAttributes":{"communicationSystemProperties":[{"name":"Business System","value":"MDI","correlationIds":["SAP_COM_0659"]},{"name":"Logical System","value":"MDI_BUPA","correlationIds":["SAP_COM_0008"]},{"name":"Business System","value":"%s","correlationIds":["SAP_COM_0008"]}],"outboundServicesProperties":[{"name":"Business Partner - Replicate from SAP S/4HANA Cloud to Client","path":"/businesspartner/v0/soap/BusinessPartnerBulkReplicateRequestIn?tenantId=%s","correlationIds":["SAP_COM_0008"],"isServiceActive":true,"additionalProperties":[{"name":"REPLICATION MODEL NAME","value":"BPMDI_CIM"},{"name":"REPLICATION_MODE","value":"C"},{"name":"OUTPUT_MODE","value":"D"}]},{"name":"Replicate Customers from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Suppliers from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Company Addresses from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Workplace Addresses from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Replicate Personal Addresses from S/4 System to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Business Partner Relationship - Replicate from SAP S/4HANA Cloud to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"Business Partner - Send Confirmation from SAP S/4HANA Cloud to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false},{"name":"BP Relationship - Send Confirmation from SAP S/4HANA Cloud to Client","correlationIds":["SAP_COM_0008"],"isServiceActive":false}]}}}`, businessSystemID, businessSystemID, receiverSubdomain, receiverSubdomain)

			t, err := template.New("").Parse(respBody)
			if err != nil {
				log.C(ctx).Errorf("An error occurred while parsing template: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while parsing MDI 'write' template: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			res := new(bytes.Buffer)
			if err = t.Execute(res, data); err != nil {
				log.C(ctx).Errorf("An error occurred while executing template: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while executing MDI 'write' template: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			log.C(ctx).Infof("Successfully processed tenant mapping notification about S4")
			if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, res.String()); statusAPIErr != nil {
				log.C(ctx).Error(statusAPIErr)
			}
			return
		}

		if tm.AssignedTenant.ApplicationNamespace == s4AppNamespace && tm.ReceiverTenant.ApplicationNamespace == billingAppNamespace {
			var inboundCert string
			isAssignedTntCfgNonEmpty := isConfigNonEmpty(string(tm.AssignedTenant.Configuration))
			if tm.Context.Operation == AssignOperation && tm.AssignedTenant.State == "CONFIG_PENDING" && tm.AssignedTenant.Configuration != nil && isAssignedTntCfgNonEmpty {
				log.C(ctx).Infof("Notification request is received for %q operation with CONFIG_PENDING state and non empty configuration. Checking for inbound certificate...", AssignOperation)
				cert := gjson.GetBytes(tm.AssignedTenant.Configuration, "credentials.inboundCommunication.oauth2mtls.certificate").String()
				if cert == "" {
					log.C(ctx).Error("The OAuth2 mTLS certificate in the assigned tenant configuration cannot be empty")
					reqBody := "{\"state\":\"CREATE_ERROR\", \"error\": \"The OAuth2 mTLS certificate in the assigned tenant configuration cannot be empty\"}"
					if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
						log.C(ctx).Error(statusAPIErr)
					}
					return
				}
				inboundCert = cert
			}

			billingSvcInstanceName := billingCatalogName + "-instance-" + formationID
			billingSvcInstanceName = truncateString(ctx, billingSvcInstanceName, serviceInstanceNameMaxLength)

			if tm.Context.Operation == UnassignOperation {
				log.C(ctx).Info("Handle subscription billing instance deletion...")
				billingSvcInstanceID, err := h.retrieveServiceInstanceIDByName(ctx, billingSvcInstanceName, tenantID)
				if err != nil {
					log.C(ctx).Error(err)
					reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
					if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
						log.C(ctx).Error(statusAPIErr)
					}
					return
				}

				if billingSvcInstanceID != "" {
					if err := h.deleteServiceKeys(ctx, billingSvcInstanceID, billingSvcInstanceName, tenantID); err != nil {
						log.C(ctx).Error(err)
						reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while deleting service key(s) for subscription billing instance: %s\"}", err.Error())
						if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
							log.C(ctx).Error(statusAPIErr)
						}
						return
					}
					if err := h.deleteServiceInstance(ctx, billingSvcInstanceID, billingSvcInstanceName, tenantID); err != nil {
						log.C(ctx).Error(err)
						reqBody := fmt.Sprintf("{\"state\":\"DELETE_ERROR\", \"error\": \"An error occurred while deleting subscription billing instance: %s\"}", err.Error())
						if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
							log.C(ctx).Error(statusAPIErr)
						}
						return
					}
				}

				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, readyResp); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			log.C(ctx).Info("Handle subscription billing instance creation...")
			billingOfferingID, err := h.retrieveServiceOffering(ctx, billingCatalogName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving service instances: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			billingPlanID, err := h.retrieveServicePlan(ctx, billingPlanName, billingOfferingID, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving service plans: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			billingSvcInstanceID, err := h.createServiceInstance(ctx, billingSvcInstanceName, billingPlanID, "", tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while creating subscription billing service instance: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			if inboundCert == "" {
				log.C(ctx).Error("The inbound certificate cannot be empty")
				reqBody := "{\"state\":\"CREATE_ERROR\", \"error\": \"The inbound certificate cannot be empty\"}"
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}
			inboundCert = removeCertificateTags(inboundCert)

			billingSvcKeyName := billingSvcInstanceName + "-key"
			billingSvcKeyName = truncateString(ctx, billingSvcKeyName, serviceBindingMaxLengthName)
			billingSvcKeyParams := fmt.Sprintf("{\"xsuaa\":{\"credential-type\":\"x509\",\"x509\":{\"certificate\":\"%s\",\"certificate-pinning\":false}}}", inboundCert)

			billingServiceKeyID, err := h.createServiceKey(ctx, billingSvcKeyName, billingSvcKeyParams, billingSvcInstanceID, billingSvcInstanceName, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while creating subscription billing service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			billingServiceKey, err := h.retrieveServiceKeyByID(ctx, billingServiceKeyID, tenantID)
			if err != nil {
				log.C(ctx).Error(err)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while retrieving MDI 'write' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			if len(billingServiceKey.Credentials) < 0 {
				log.C(ctx).Errorf("The credentials for MDI 'write' service key with ID: %q should not be empty", billingServiceKey.ID)
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"The service key for the subscription billing instance shoud not be empty: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			data, err := h.buildTemplateDataForSubscriptionBilling(billingServiceKey.Credentials)
			if err != nil {
				log.C(ctx).Errorf("An error occurred while building template data with the MDI 'write' service key: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while building template data with the MDI 'write' service key: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			respBody := fmt.Sprintf("{\"state\":\"CONFIG_PENDING\",\"configuration\":{\"credentials\":{\"outboundCommunication\":{\"oauth2mtls\":{\"url\":\"{{ .URL }}\",\"tokenServiceUrl\":\"{{ .TokenURL }}\",\"clientId\":\"{{ .ClientID }}\",\"certificate\":\"%s\",\"correlationIds\":[\"SAP_COM_0642\"]}}}}}", inboundCert)
			t, err := template.New("").Parse(respBody)
			if err != nil {
				log.C(ctx).Errorf("An error occurred while parsing template: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while parsing subscription billing template: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			res := new(bytes.Buffer)
			if err = t.Execute(res, data); err != nil {
				log.C(ctx).Errorf("An error occurred while executing subscription billing template: %s", err.Error())
				reqBody := fmt.Sprintf("{\"state\":\"CREATE_ERROR\", \"error\": \"An error occurred while executing subscription billing template: %s\"}", err.Error())
				if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, reqBody); statusAPIErr != nil {
					log.C(ctx).Error(statusAPIErr)
				}
				return
			}

			log.C(ctx).Infof("Successfully processed tenant mapping notification for subscription billing")
			if statusAPIErr := h.sendStatusAPIRequest(ctx, statusAPIURL, res.String()); statusAPIErr != nil {
				log.C(ctx).Error(statusAPIErr)
			}
			return
		}
	}(h.mu)
}

func removeCertificateTags(cert string) string {
	cert = strings.TrimPrefix(cert, beginCertTag)
	cert = strings.TrimSuffix(cert, endCertTag)
	cert = strings.Replace(cert, "\n", "", -1)

	return cert
}

func truncateString(ctx context.Context, str string, maxLength int) string {
	if len(str) > maxLength {
		log.C(ctx).Infof("The length of the string is bigger than %d, truncating it...", maxLength)
		str = str[:maxLength]
		return str
	}

	return str
}

func (h *Handler) sendStatusAPIRequest(ctx context.Context, statusAPIURL, reqBody string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, statusAPIURL, bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return errors.Wrapf(err, "An error occurred while building status API request")
	}
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)

	log.C(ctx).Infof("Sending notification status response to the status API URL: %s ...", statusAPIURL)
	resp, err := h.mtlsHTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "An error occurred while executing request to the status API")
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "An error occurred while reading status API response")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Wrapf(err, "An error occurred while calling UCL status API. Received status: %d and body: %s", resp.StatusCode, body)
	}

	return nil
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

func (h *Handler) retrieveServiceOffering(ctx context.Context, catalogName, tenantID string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceOfferingsPath, SubaccountKey, tenantID)
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

func (h *Handler) retrieveServicePlan(ctx context.Context, planName, offeringID, tenantID string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServicePlansPath, SubaccountKey, tenantID)
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

func (h *Handler) createServiceInstance(ctx context.Context, serviceInstanceName, planID, parameters, tenantID string) (string, error) {
	siReqBody := &types.ServiceInstanceReqBody{
		Name:          serviceInstanceName,
		ServicePlanId: planID,
		Parameters:    json.RawMessage(parameters),
	}

	siReqBodyBytes, err := json.Marshal(siReqBody)
	if err != nil {
		return "", errors.Errorf("Failed to marshal service instance body: %v", err)
	}

	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceInstancesPath, SubaccountKey, tenantID)
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

		opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, tenantID)
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

func (h *Handler) deleteServiceKeys(ctx context.Context, serviceInstanceID, serviceInstanceName, tenantID string) error {
	svcKeyIDs, err := h.retrieveServiceKeyIDsByInstanceID(ctx, serviceInstanceID, serviceInstanceName, tenantID)
	if err != nil {
		return err
	}

	for _, keyID := range svcKeyIDs {
		svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", keyID)
		strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, svcKeyPath, SubaccountKey, tenantID)
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

			opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, tenantID)
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

func (h *Handler) deleteServiceInstance(ctx context.Context, serviceInstanceID, serviceInstanceName, tenantID string) error {
	svcInstancePath := paths.ServiceInstancesPath + fmt.Sprintf("/%s", serviceInstanceID)
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, svcInstancePath, SubaccountKey, tenantID)
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

		opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, tenantID)
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

func (h *Handler) retrieveServiceInstanceIDByName(ctx context.Context, serviceInstanceName, tenantID string) (string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceInstancesPath, SubaccountKey, tenantID)
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

func (h *Handler) createServiceKey(ctx context.Context, serviceKeyName, serviceKeyParams, serviceInstanceID, serviceInstanceName, tenantID string) (string, error) {
	serviceKeyReqBody := &types.ServiceKeyReqBody{}
	if strings.Contains(serviceInstanceName, billingCatalogName) {
		serviceKeyReqBody = &types.ServiceKeyReqBody{
			Name:              serviceKeyName,
			ServiceInstanceId: serviceInstanceID,
			Parameters:        json.RawMessage(serviceKeyParams),
		}
	} else {
		serviceKeyReqBody = &types.ServiceKeyReqBody{
			Name:              serviceKeyName,
			ServiceInstanceId: serviceInstanceID,
		}
	}

	serviceKeyReqBodyBytes, err := json.Marshal(serviceKeyReqBody)
	if err != nil {
		return "", errors.Errorf("Failed to marshal service key body: %v", err)
	}

	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceBindingsPath, SubaccountKey, tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service bindings URL")
	}

	log.C(ctx).Infof("Creating service key for service instance with name: %q", serviceInstanceName)
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

		opURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, opStatusPath, SubaccountKey, tenantID)
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

	log.C(ctx).Infof("Successfully create service key for service instance with name: %q synchronously", serviceInstanceName)
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

func (h *Handler) retrieveServiceKeyIDsByInstanceID(ctx context.Context, serviceInstanceID, serviceInstanceName, tenantID string) ([]string, error) {
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, paths.ServiceBindingsPath, SubaccountKey, tenantID)
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

func (h *Handler) retrieveServiceKeyByID(ctx context.Context, serviceKeyID, tenantID string) (*types.ServiceKey, error) {
	svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", serviceKeyID)
	strURL, err := buildURL(h.cfg.ServiceManagerCfg.URL, svcKeyPath, SubaccountKey, tenantID)
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
	tokenPath := "/oauth/token"

	data := map[string]string{
		"SystemID":     systemID,
		"URL":          svcKeyURI,
		"TokenURL":     svcKeyTokenURL + tokenPath,
		"ClientID":     svcKeyClientID,
		"ClientSecret": svcKeyClientSecret,
	}

	return data, nil
}

func (h *Handler) buildTemplateDataForSubscriptionBilling(serviceKeyCredentials json.RawMessage) (map[string]string, error) {
	svcURL, ok := gjson.Get(string(serviceKeyCredentials), "url").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'url' property")
	}

	svcKeyClientID, ok := gjson.Get(string(serviceKeyCredentials), "uaa.clientid").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uaa.clientid' property")
	}

	svcKeyTokenURL, ok := gjson.Get(string(serviceKeyCredentials), "uaa.url").Value().(string)
	if !ok {
		return nil, errors.New("could not find 'uaa.url' property")
	}
	tokenPath := "/oauth/token"

	data := map[string]string{
		"URL":      svcURL,
		"TokenURL": svcKeyTokenURL + tokenPath,
		"ClientID": svcKeyClientID,
	}

	return data, nil
}

func isConfigNonEmpty(configuration string) bool {
	if configuration != "" && configuration != "{}" && configuration != "\"\"" && configuration != "null" {
		return true
	}

	return false
}
