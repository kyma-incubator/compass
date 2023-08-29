package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/paths"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/config"
	"github.com/pkg/errors"
)

const (
	// SubaccountKey is used as URL parameter
	SubaccountKey = "subaccount_id"
	// LocationHeaderKey is used in the async API calls
	LocationHeaderKey = "Location"
)

type client struct {
	cfg            *config.Config
	callerProvider *CallerProvider
}

// NewClient creates a new client
func NewClient(cfg *config.Config, callerProvider *CallerProvider) *client {
	return &client{
		cfg:            cfg,
		callerProvider: callerProvider,
	}
}

// RetrieveServiceOffering retrieves a Service Offering
func (c *client) RetrieveServiceOffering(ctx context.Context, region, catalogName, subaccountID string) (string, error) {
	strURL, err := buildURL(c.cfg.InstanceSMURLPath, paths.ServiceOfferingsPath, SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service offerings URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Listing service offerings for subaccount with ID: %q...", subaccountID)
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return "", errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("failed to read service offerings response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to get service offerings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service offerings for subaccount with ID: %q...", subaccountID)

	var offerings types.ServiceOfferings
	err = json.Unmarshal(body, &offerings)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal service offerings: %v", err)
	}

	var offeringID string
	for _, item := range offerings.Items {
		if item.CatalogName == catalogName {
			offeringID = item.ID
			break
		}
	}

	if offeringID == "" {
		return "", errors.Errorf("couldn't find service offering for catalog name: %s", catalogName)
	}

	log.C(ctx).Infof("Service offering with ID: %q for catalog name: %q and subaccount: %q is found", offeringID, catalogName, subaccountID)

	return offeringID, nil
}

// RetrieveServicePlan retrieves a Service Plan
func (c *client) RetrieveServicePlan(ctx context.Context, region, planName, offeringID, subaccountID string) (string, error) {
	strURL, err := buildURL(c.cfg.InstanceSMURLPath, paths.ServicePlansPath, SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service plans URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Listing service plans for subaccount with ID: %q...", subaccountID)
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return "", errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("failed to read service plans response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to get service plans, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service plans for subaccount with ID: %q...", subaccountID)

	var plans types.ServicePlans
	err = json.Unmarshal(body, &plans)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal service plans: %v", err)
	}

	var planID string
	for _, item := range plans.Items {
		if item.CatalogName == planName && item.ServiceOfferingId == offeringID {
			planID = item.ID
			break
		}
	}

	if planID == "" {
		return "", errors.Errorf("couldn't find service plan for catalog name: %s and offering ID: %s", planName, offeringID)
	}

	log.C(ctx).Infof("Service plan with ID: %q for offering with ID: %q and subaccount: %q is found", planID, offeringID, subaccountID)

	return planID, nil
}

// RetrieveServiceKeyByID retrieves a Service Key by its ID
func (c *client) RetrieveServiceKeyByID(ctx context.Context, region, serviceKeyID, subaccountID string) (*types.ServiceKey, error) {
	svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", serviceKeyID)
	strURL, err := buildURL(c.cfg.InstanceSMURLPath, svcKeyPath, SubaccountKey, subaccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building service binding URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Getting service key by ID: %s for subaccount with ID: %q", serviceKeyID, subaccountID)
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read service binding response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get service bindings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service key by ID: %s for subaccount with ID: %q", serviceKeyID, subaccountID)

	var serviceKey types.ServiceKey
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return nil, errors.Errorf("failed to unmarshal service key: %v", err)
	}

	return &serviceKey, nil
}

// RetrieveServiceInstanceIDByName retrieves a Service Instance by its name
func (c *client) RetrieveServiceInstanceIDByName(ctx context.Context, region, serviceInstanceName, subaccountID string) (string, error) {
	strURL, err := buildURL(c.cfg.InstanceSMURLPath, paths.ServiceInstancesPath, SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Listing service instances for subaccount with ID: %s...", subaccountID)
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return "", errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("failed to read service instances response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to get service instances, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service instances for subaccount with ID: %q", subaccountID)

	var instances types.ServiceInstances
	err = json.Unmarshal(body, &instances)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal service instances: %v", err)
	}

	var instanceID string
	for _, item := range instances.Items {
		if item.Name == serviceInstanceName {
			instanceID = item.ID
			break
		}
	}

	if instanceID == "" {
		log.C(ctx).Warnf("No instance ID found by name: %q for subaccount with ID: %q", serviceInstanceName, subaccountID)
		return "", nil
	}

	log.C(ctx).Infof("Successfully found service instance with ID: %q by instance name: %q for subaccount with ID: %q", instanceID, serviceInstanceName, subaccountID)
	return instanceID, nil
}

// CreateServiceInstance creates a Service Instance both synchronously and asynchronously
func (c *client) CreateServiceInstance(ctx context.Context, region, serviceInstanceName, planID, subaccountID string, parameters []byte) (string, error) {
	siReqBody := &types.ServiceInstanceReqBody{
		Name:          serviceInstanceName,
		ServicePlanID: planID,
		Parameters:    parameters, // todo::: most probably should be provided as `parameters` label in the TM notification body - `receiverTenant.parameters`?
	}

	siReqBodyBytes, err := json.Marshal(siReqBody)
	if err != nil {
		return "", errors.Errorf("failed to marshal service instance body: %v", err)
	}

	strURL, err := buildURL(c.cfg.InstanceSMURLPath, paths.ServiceInstancesPath, SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(siReqBodyBytes))
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Creating service instance with name: %q from plan with ID: %q and subaccount ID: %q", serviceInstanceName, planID, subaccountID)
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return "", errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("failed to read response body from service instance creation request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", errors.Errorf("failed to create service instance, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous service instance creation...")
		opStatusPath := resp.Header.Get(LocationHeaderKey)
		if opStatusPath == "" {
			return "", errors.Errorf("the service instance operation status path from %s header should not be empty", LocationHeaderKey)
		}

		opURL, err := buildURL(c.cfg.InstanceSMURLPath, opStatusPath, SubaccountKey, subaccountID)
		if err != nil {
			return "", errors.Wrapf(err, "while building asynchronous service instance operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return "", err
		}

		ticker := time.NewTicker(c.cfg.Ticker)
		timeout := time.After(c.cfg.Timeout)
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service instance with name: %q and subaccount with ID: %q", serviceInstanceName, subaccountID)
				opResp, err := caller.Call(opReq)
				if err != nil {
					return "", err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := ioutil.ReadAll(opResp.Body)
				if err != nil {
					return "", errors.Errorf("failed to read operation response body from asynchronous service instance creation request: %v", err)
				}

				if opResp.StatusCode != http.StatusOK {
					return "", errors.Errorf("failed to get asynchronous service instance operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
				}

				var opStatus types.OperationStatus
				err = json.Unmarshal(opBody, &opStatus)
				if err != nil {
					return "", errors.Errorf("failed to unmarshal service instance operation status: %v", err)
				}

				if opStatus.State == types.OperationStateInProgress {
					log.C(ctx).Infof("The asynchronous service instance operation state is still: %q", types.OperationStateInProgress)
					continue
				}

				if opStatus.State != types.OperationStateSucceeded {
					return "", errors.Errorf("The asynchronous service instance operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
				}

				log.C(ctx).Infof("The asynchronous operation status for service instance with name: %q and subaccount: %q finished with state: %s", serviceInstanceName, subaccountID, opStatus.State)
				serviceInstanceID := opStatus.ResourceID
				if serviceInstanceID == "" {
					return "", errors.New("the service instance ID could not be empty")
				}

				return serviceInstanceID, nil
			case <-timeout:
				return "", errors.New("Timeout waiting for asynchronous operation status to finish")
			}
		}
	}

	var serviceInstance types.ServiceInstance
	err = json.Unmarshal(body, &serviceInstance)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal service instance: %v", err)
	}

	serviceInstanceID := serviceInstance.ID
	if serviceInstanceID == "" {
		return "", errors.New("the service instance ID could not be empty")
	}
	log.C(ctx).Infof("Successfully created service instance with name: %q, ID: %q and subaccount ID: %q synchronously", serviceInstanceName, serviceInstanceID, subaccountID)

	return serviceInstanceID, nil
}

// CreateServiceKey creates a Service Key both synchronously and asynchronously
func (c *client) CreateServiceKey(ctx context.Context, region, serviceKeyName, serviceInstanceID, subaccountID string, parameters []byte) (string, error) {
	serviceKeyReqBody := &types.ServiceKeyReqBody{
		Name:              serviceKeyName,
		ServiceInstanceID: serviceInstanceID,
		Parameters:        parameters,
	}

	serviceKeyReqBodyBytes, err := json.Marshal(serviceKeyReqBody)
	if err != nil {
		return "", errors.Errorf("failed to marshal service key body: %v", err)
	}

	strURL, err := buildURL(c.cfg.InstanceSMURLPath, paths.ServiceBindingsPath, SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building service bindings URL")
	}

	log.C(ctx).Infof("Creating service key for service instance with ID: %q and subaccount: %q", serviceInstanceID, subaccountID)
	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(serviceKeyReqBodyBytes))
	if err != nil {
		return "", err
	}

	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return "", errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		return "", err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Errorf("failed to read response body from service key creation request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", errors.Errorf("failed to create service key, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous service key creation...")
		opStatusPath := resp.Header.Get(LocationHeaderKey)
		if opStatusPath == "" {
			return "", errors.Errorf("the service key operation status path from %s header should not be empty", LocationHeaderKey)
		}

		opURL, err := buildURL(c.cfg.InstanceSMURLPath, opStatusPath, SubaccountKey, subaccountID)
		if err != nil {
			return "", errors.Wrapf(err, "while building asynchronous service key operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return "", err
		}

		ticker := time.NewTicker(c.cfg.Ticker)
		timeout := time.After(c.cfg.Timeout)
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service key with name: %q and subaccount: %q", serviceKeyName, subaccountID)
				opResp, err := caller.Call(opReq)
				if err != nil {
					return "", err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := ioutil.ReadAll(opResp.Body)
				if err != nil {
					return "", errors.Errorf("failed to read operation response body from asynchronous service key creation request: %v", err)
				}

				if opResp.StatusCode != http.StatusOK {
					return "", errors.Errorf("failed to get asynchronous service key operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
				}

				var opStatus types.OperationStatus
				err = json.Unmarshal(opBody, &opStatus)
				if err != nil {
					return "", errors.Errorf("failed to unmarshal service key operation status: %v", err)
				}

				if opStatus.State == types.OperationStateInProgress {
					log.C(ctx).Infof("The asynchronous service key operation state is still: %q", types.OperationStateInProgress)
					continue
				}

				if opStatus.State != types.OperationStateSucceeded {
					return "", errors.Errorf("the asynchronous service key operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
				}

				log.C(ctx).Infof("The asynchronous operation status for service key with name: %q finished with state: %q", serviceKeyName, opStatus.State)
				serviceKeyID := opStatus.ResourceID
				if serviceKeyID == "" {
					return "", errors.New("the service key ID could not be empty")
				}

				return serviceKeyID, nil
			case <-timeout:
				return "", errors.New("timeout waiting for asynchronous operation status to finish")
			}
		}
	}

	var serviceKey types.ServiceKey
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal service key: %v", err)
	}

	serviceKeyID := serviceKey.ID
	if serviceKeyID == "" {
		return "", errors.New("the service key ID could not be empty")
	}
	log.C(ctx).Infof("Successfully created service key with name: %q and subaccount: %q synchronously", serviceKeyName, subaccountID)

	return serviceKeyID, nil
}

// DeleteServiceInstance deletes a Service Instance both synchronously and asynchronously
func (c *client) DeleteServiceInstance(ctx context.Context, region, serviceInstanceID, serviceInstanceName, subaccountID string) error {
	svcInstancePath := paths.ServiceInstancesPath + fmt.Sprintf("/%s", serviceInstanceID)
	strURL, err := buildURL(c.cfg.InstanceSMURLPath, svcInstancePath, SubaccountKey, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while building service instances URL")
	}

	req, err := http.NewRequest(http.MethodDelete, strURL, nil)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Deleting service instance with ID: %q, name: %q and subaccount: %q", serviceInstanceID, serviceInstanceName, subaccountID)
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return errors.Errorf("error while getting caller for region %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("failed to read response body from service instance deletion request: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errors.Errorf("failed to delete service instance, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous service instance deletion...")
		opStatusPath := resp.Header.Get(LocationHeaderKey)
		if opStatusPath == "" {
			return errors.Errorf("the service instance operation status path from %s header should not be empty", LocationHeaderKey)
		}

		opURL, err := buildURL(c.cfg.InstanceSMURLPath, opStatusPath, SubaccountKey, subaccountID)
		if err != nil {
			return errors.Wrapf(err, "while building asynchronous service instance operation URL")
		}

		opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
		if err != nil {
			return err
		}

		ticker := time.NewTicker(c.cfg.Ticker)
		timeout := time.After(c.cfg.Timeout)
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Getting asynchronous operation status for service instance with ID: %q, name: %q and subaccount: %q", serviceInstanceID, serviceInstanceName, subaccountID)
				opResp, err := caller.Call(opReq)
				if err != nil {
					return err
				}
				defer closeResponseBody(ctx, opResp)

				opBody, err := ioutil.ReadAll(opResp.Body)
				if err != nil {
					return errors.Errorf("failed to read operation response body from asynchronous service instance deletion request: %v", err)
				}

				if opResp.StatusCode != http.StatusOK {
					return errors.Errorf("failed to get asynchronous service instance operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
				}

				var opStatus types.OperationStatus
				err = json.Unmarshal(opBody, &opStatus)
				if err != nil {
					return errors.Errorf("failed to unmarshal service instance operation status: %v", err)
				}

				if opStatus.State == types.OperationStateInProgress {
					log.C(ctx).Infof("The asynchronous service instance operation state is still: %q", types.OperationStateInProgress)
					continue
				}

				if opStatus.State != types.OperationStateSucceeded {
					return errors.Errorf("the asynchronous service instance operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
				}

				log.C(ctx).Infof("The asynchronous operation status for service instance with name: %q finished with state: %q", serviceInstanceName, opStatus.State)
				return nil
			case <-timeout:
				return errors.New("timeout waiting for asynchronous operation status to finish")
			}
		}
	}

	log.C(ctx).Infof("Successfully deleted service instance with ID: %q and subaccount: %q synchronously", serviceInstanceID, subaccountID)

	return nil
}

// DeleteServiceKeys deletes all Service Keys related to a Service Instance both synchronously and asynchronously
func (c *client) DeleteServiceKeys(ctx context.Context, region, serviceInstanceID, serviceInstanceName, subaccountID string) error {
	caller, err := c.callerProvider.GetCaller(*c.cfg, region)
	if err != nil {
		return errors.Errorf("error while getting caller for region %s", region)
	}

	svcKeyIDs, err := c.retrieveServiceKeysIDByInstanceID(ctx, caller, serviceInstanceID, serviceInstanceName, subaccountID)
	if err != nil {
		return err
	}

	for _, keyID := range svcKeyIDs {
		svcKeyPath := paths.ServiceBindingsPath + fmt.Sprintf("/%s", keyID)
		strURL, err := buildURL(c.cfg.InstanceSMURLPath, svcKeyPath, SubaccountKey, subaccountID)
		if err != nil {
			return errors.Wrapf(err, "while building service binding URL")
		}

		req, err := http.NewRequest(http.MethodDelete, strURL, nil)
		if err != nil {
			return err
		}

		log.C(ctx).Infof("Deleting service binding with ID: %q for subaccount: %q", keyID, subaccountID)
		resp, err := caller.Call(req)
		if err != nil {
			return err
		}
		defer closeResponseBody(ctx, resp)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Errorf("failed to read response body from service binding deletion request: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			return errors.Errorf("failed to delete service binding, status: %d, body: %s", resp.StatusCode, body)
		}

		if resp.StatusCode == http.StatusAccepted {
			log.C(ctx).Infof("Handle asynchronous service binding deletion...")
			opStatusPath := resp.Header.Get(LocationHeaderKey)
			if opStatusPath == "" {
				return errors.Errorf("the service binding operation status path from %s header should not be empty", LocationHeaderKey)
			}

			opURL, err := buildURL(c.cfg.InstanceSMURLPath, opStatusPath, SubaccountKey, subaccountID)
			if err != nil {
				return errors.Wrapf(err, "while building asynchronous service binding operation URL")
			}

			opReq, err := http.NewRequest(http.MethodGet, opURL, nil)
			if err != nil {
				return err
			}

			ticker := time.NewTicker(c.cfg.Ticker)
			timeout := time.After(c.cfg.Timeout)
			for {
				select {
				case <-ticker.C:
					log.C(ctx).Infof("Getting asynchronous operation status for service binding with ID: %q and subaccount: %q", keyID, subaccountID)
					opResp, err := caller.Call(opReq)
					if err != nil {
						return err
					}
					defer closeResponseBody(ctx, opResp)

					opBody, err := ioutil.ReadAll(opResp.Body)
					if err != nil {
						return errors.Errorf("failed to read operation response body from asynchronous service binding deletion request: %v", err)
					}

					if opResp.StatusCode != http.StatusOK {
						return errors.Errorf("failed to get asynchronous service binding operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
					}

					var opStatus types.OperationStatus
					err = json.Unmarshal(opBody, &opStatus)
					if err != nil {
						return errors.Errorf("failed to unmarshal service binding operation status: %v", err)
					}

					if opStatus.State == types.OperationStateInProgress {
						log.C(ctx).Infof("The asynchronous service binding operation state is still: %q", types.OperationStateInProgress)
						continue
					}

					if opStatus.State != types.OperationStateSucceeded {
						return errors.Errorf("the asynchronous service binding operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
					}

					log.C(ctx).Infof("The asynchronous operation status for service binding with ID: %q finished with state: %q", keyID, opStatus.State)
					return nil
				case <-timeout:
					return errors.New("timeout waiting for asynchronous operation status to finish")
				}
			}
		}

		log.C(ctx).Infof("Successfully deleted service binding with ID: %q for subaccount: %q synchronously", keyID, subaccountID)
	}

	return nil
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

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func (c *client) retrieveServiceKeysIDByInstanceID(ctx context.Context, caller ExternalSvcCaller, serviceInstanceID, serviceInstanceName, subaccountID string) ([]string, error) {
	strURL, err := buildURL(c.cfg.InstanceSMURLPath, paths.ServiceBindingsPath, SubaccountKey, subaccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building service binding URL")
	}

	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Listing service bindings for instance with ID: %q, name: %q and subaccount: %q", serviceInstanceID, serviceInstanceName, subaccountID)
	resp, err := caller.Call(req)
	if err != nil {
		log.C(ctx).Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read service binding response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get service bindings, status: %d, body: %s", resp.StatusCode, body)
	}
	log.C(ctx).Infof("Successfully fetch service bindings for instance with ID: %q, name: %q and subaccount: %q", serviceInstanceID, serviceInstanceName, subaccountID)

	var svcKeys types.ServiceKeys
	err = json.Unmarshal(body, &svcKeys)
	if err != nil {
		return nil, errors.Errorf("Failed to unmarshal service keys: %v", err)
	}

	serviceKeysIDs := make([]string, 0, len(svcKeys.Items))
	for _, key := range svcKeys.Items {
		if key.ServiceInstanceID == serviceInstanceID {
			serviceKeysIDs = append(serviceKeysIDs, key.ID)
		}
	}
	log.C(ctx).Infof("Service instance with ID: %q and name: %q has/have %d keys(s)", serviceInstanceID, serviceInstanceName, len(serviceKeysIDs))

	return serviceKeysIDs, nil
}
