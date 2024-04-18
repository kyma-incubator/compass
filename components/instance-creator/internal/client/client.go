package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/config"
	"github.com/pkg/errors"
)

const (
	// SubaccountKey is used as URL parameter
	SubaccountKey = "subaccount_id"
	// LocationHeaderKey is used in the async API calls
	LocationHeaderKey = "Location"
	// LabelsQueryKey is used as URL parameter for Service Manager label querying
	LabelsQueryKey = "labelQuery"
)

// ExternalSvcCaller is used to call external services with given authentication
//
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCaller interface {
	Call(*http.Request) (*http.Response, error)
}

// ExternalSvcCallerProvider provides ExternalSvcCaller based on the provided config and region
//
//go:generate mockery --name=ExternalSvcCallerProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCallerProvider interface {
	GetCaller(cfg config.Config, region string) (ExternalSvcCaller, error)
}

type client struct {
	cfg            config.Config
	callerProvider ExternalSvcCallerProvider
}

// NewClient creates a new client
func NewClient(cfg config.Config, callerProvider ExternalSvcCallerProvider) *client {
	return &client{
		cfg:            cfg,
		callerProvider: callerProvider,
	}
}

// RetrieveResource retrieves a given resource from SM by some criteria
func (c *client) RetrieveResource(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) (string, error) {
	strURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, resources.GetURLPath(), SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building %s URL", resources.GetType())
	}

	log.C(ctx).Infof("Listing %s for subaccount with ID: %q...", resources.GetType(), subaccountID)
	body, err := c.executeSyncRequest(ctx, strURL, region)
	if err != nil {
		return "", errors.Wrapf(err, "while executing request for listing %s for subaccount with ID: %q", resources.GetType(), subaccountID)
	}

	log.C(ctx).Infof("Successfully listed %s for subaccount with ID: %q...", resources.GetType(), subaccountID)

	err = json.Unmarshal(body, &resources)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal %s: %v", resources.GetType(), err)
	}

	resourceID, err := resourceMatchParams.Match(resources)
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("%s record with ID: %q and subaccount: %q is found", resources.GetType(), resourceID, subaccountID)

	return resourceID, nil
}

// RetrieveResourceByID retrieves a given resource from SM by its ID
func (c *client) RetrieveResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource) (resources.Resource, error) {
	resourcePath := resource.GetResourceURLPath() + fmt.Sprintf("/%s", resource.GetResourceID())
	strURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, resourcePath, SubaccountKey, subaccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building %s URL", resource.GetResourceType())
	}

	log.C(ctx).Infof("Getting %s by ID: %s for subaccount with ID: %q", resource.GetResourceType(), resource.GetResourceID(), subaccountID)
	body, err := c.executeSyncRequest(ctx, strURL, region)
	if err != nil {
		return nil, errors.Wrapf(err, "while executing request for getting %s for subaccount with ID: %q", resource.GetResourceType(), subaccountID)
	}
	log.C(ctx).Infof("Successfully got %s by ID: %s for subaccount with ID: %q", resource.GetResourceType(), resource.GetResourceID(), subaccountID)

	err = json.Unmarshal(body, &resource)
	if err != nil {
		return nil, errors.Errorf("failed to unmarshal %s: %v", resource.GetResourceType(), err)
	}

	return resource, nil
}

// RetrieveMultipleResourcesIDsByLabels retrieves resources IDs from SM by labels
func (c *client) RetrieveMultipleResourcesIDsByLabels(ctx context.Context, region, subaccountID string, resources resources.Resources, labels map[string][]string) ([]string, error) {
	strURL, err := buildURLWithLabels(c.cfg.RegionToInstanceConfig[region].SMURL, resources.GetURLPath(), SubaccountKey, subaccountID, LabelsQueryKey, labels)
	if err != nil {
		return nil, errors.Wrapf(err, "while building %s URL with labels", resources.GetType())
	}

	log.C(ctx).Infof("Listing %s by labels: %v for subaccount with ID: %q", resources.GetType(), labels, subaccountID)
	body, err := c.executeSyncRequest(ctx, strURL, region)
	if err != nil {
		return nil, errors.Wrapf(err, "while executing request for listing %s for subaccount with ID: %q", resources.GetType(), subaccountID)
	}
	log.C(ctx).Infof("Successfully got %s by labels: %v for subaccount with ID: %q", resources.GetType(), labels, subaccountID)

	err = json.Unmarshal(body, &resources)
	if err != nil {
		return nil, errors.Errorf("failed to unmarshal %s: %v", resources.GetType(), err)
	}

	return resources.GetIDs(), nil
}

// RetrieveRawResourceByID retrieves a given resource from SM by its ID and returns the raw json body response
func (c *client) RetrieveRawResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource) (json.RawMessage, error) {
	resourcePath := resource.GetResourceURLPath() + fmt.Sprintf("/%s", resource.GetResourceID())
	strURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, resourcePath, SubaccountKey, subaccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building %s URL", resource.GetResourceType())
	}

	log.C(ctx).Infof("Getting %s by ID: %s for subaccount with ID: %q", resource.GetResourceType(), resource.GetResourceID(), subaccountID)
	body, err := c.executeSyncRequest(ctx, strURL, region)
	if err != nil {
		return nil, errors.Wrapf(err, "while executing request for getting %s for subaccount with ID: %q", resource.GetResourceType(), subaccountID)
	}
	log.C(ctx).Infof("Successfully got %s by ID: %s for subaccount with ID: %q", resource.GetResourceType(), resource.GetResourceID(), subaccountID)

	return body, nil
}

// CreateResource creates a given resource in SM
func (c *client) CreateResource(ctx context.Context, region, subaccountID string, resourceReqBody resources.ResourceRequestBody, resource resources.Resource) (string, error) {
	resourceReqBodyBytes, err := json.Marshal(resourceReqBody)
	if err != nil {
		return "", errors.Errorf("failed to marshal %s body: %v", resource.GetResourceType(), err)
	}

	strURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, resource.GetResourceURLPath(), SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building %s URL", resource.GetResourceType())
	}

	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(resourceReqBodyBytes))
	if err != nil {
		return "", err
	}

	log.C(ctx).Infof("Creating %s for subaccount with ID: %q", resource.GetResourceType(), subaccountID)
	caller, err := c.callerProvider.GetCaller(c.cfg, region)
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
		return "", errors.Errorf("failed to read response body from %s creation request: %v", resource.GetResourceType(), err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", errors.Errorf("failed to create record of %s, status: %d, body: %s", resource.GetResourceType(), resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous creation of %s...", resource.GetResourceType())
		resourceID, err := c.executeAsyncRequest(ctx, resp, caller, region, subaccountID, true)
		if err != nil {
			return "", errors.Wrapf(err, "while handling asynchronous creation of %s in subaccount with ID: %q", resource.GetResourceType(), subaccountID)
		}
		if resourceID == "" {
			return "", errors.Errorf("the %s ID could not be empty", resource.GetResourceType())
		}

		return resourceID, nil
	}

	err = json.Unmarshal(body, &resource)
	if err != nil {
		return "", errors.Errorf("failed to unmarshal %s: %v", resource.GetResourceType(), err)
	}

	resourceID := resource.GetResourceID()
	if resourceID == "" {
		return "", errors.Errorf("the %s ID could not be empty", resource.GetResourceType())
	}
	log.C(ctx).Infof("Successfully created %s for subaccount with ID: %q synchronously", resource.GetResourceType(), subaccountID)

	return resourceID, nil
}

// DeleteResource deletes a given resource from SM by its ID
func (c *client) DeleteResource(ctx context.Context, region, subaccountID string, resource resources.Resource) error {
	resourcePath := resource.GetResourceURLPath() + fmt.Sprintf("/%s", resource.GetResourceID())
	strURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, resourcePath, SubaccountKey, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while building %s URL", resource.GetResourceType())
	}

	req, err := http.NewRequest(http.MethodDelete, strURL, nil)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Deleting %s with ID: %q and subaccount: %q", resource.GetResourceType(), resource.GetResourceID(), subaccountID)
	caller, err := c.callerProvider.GetCaller(c.cfg, region)
	if err != nil {
		return errors.Errorf("error while getting caller for region: %s", region)
	}

	resp, err := caller.Call(req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("failed to read response body from %s deletion request: %v", resource.GetResourceType(), err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errors.Errorf("failed to delete %s, status: %d, body: %s", resource.GetResourceType(), resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusAccepted {
		log.C(ctx).Infof("Handle asynchronous %s deletion...", resource.GetResourceType())
		_, err := c.executeAsyncRequest(ctx, resp, caller, region, subaccountID, false)
		if err != nil {
			return errors.Wrapf(err, "while deleting %s with ID: %q and subaccount: %q", resource.GetResourceType(), resource.GetResourceID(), subaccountID)
		}
		return nil
	}

	log.C(ctx).Infof("Successfully deleted %s with ID: %q and subaccount: %q synchronously", resource.GetResourceType(), resource.GetResourceID(), subaccountID)

	return nil
}

// DeleteMultipleResources deletes multiple resources from SM by some criteria
func (c *client) DeleteMultipleResources(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) error {
	resourceIDs, err := c.RetrieveMultipleResources(ctx, region, subaccountID, resources, resourceMatchParams)
	if err != nil {
		return err
	}

	for _, resourceID := range resourceIDs {
		resource, err := c.prepareResourceForDeletion(resources.GetType(), resourceID)
		if err != nil {
			return err
		}
		err = c.DeleteResource(ctx, region, subaccountID, resource)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteMultipleResourcesByIDs deletes multiple resources from SM by IDs
func (c *client) DeleteMultipleResourcesByIDs(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceIDs []string) error {
	for _, resourceID := range resourceIDs {
		resource, err := c.prepareResourceForDeletion(resources.GetType(), resourceID)
		if err != nil {
			return err
		}
		err = c.DeleteResource(ctx, region, subaccountID, resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) prepareResourceForDeletion(resourceType, resourceID string) (resources.Resource, error) {
	switch resourceType {
	case types.ServiceOfferingsType:
		return &types.ServiceOffering{ID: resourceID}, nil
	case types.ServicePlansType:
		return &types.ServicePlan{ID: resourceID}, nil
	case types.ServiceInstancesType:
		return &types.ServiceInstance{ID: resourceID}, nil
	case types.ServiceBindingsType:
		return &types.ServiceBinding{ID: resourceID}, nil
	default:
		return nil, errors.New("unknown resource type")
	}
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

func buildURLWithLabels(baseURL, path, tenantKey, tenantValue, labelsKey string, labelsValue map[string][]string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Path params
	base.Path += path

	// Build labels query
	operators := make([]string, 0, len(labelsValue))
	for key, values := range labelsValue {
		operators = append(operators, fmt.Sprintf("%s in ('%s')", key, strings.Join(values, "', '")))
	}
	labelsQuery := strings.Join(operators, " and ")

	// Query params
	params := url.Values{}
	params.Add(tenantKey, tenantValue)
	params.Add(labelsKey, labelsQuery)

	base.RawQuery = params.Encode()

	return base.String(), nil
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func (c *client) executeSyncRequest(ctx context.Context, strURL, region string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	caller, err := c.callerProvider.GetCaller(c.cfg, region)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting caller for region: %s", region)
	}

	resp, err := sendRequestWithRetry(ctx, caller, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read object, response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get object(s), status: %d, body: %s", resp.StatusCode, body)
	}

	return body, nil
}

func (c *client) executeAsyncRequest(ctx context.Context, resp *http.Response, caller ExternalSvcCaller, region, subaccountID string, isCreateRequest bool) (string, error) {
	opStatusPath := resp.Header.Get(LocationHeaderKey)
	if opStatusPath == "" {
		return "", errors.Errorf("the operation status path from %s header should not be empty", LocationHeaderKey)
	}

	opURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, opStatusPath, SubaccountKey, subaccountID)
	if err != nil {
		return "", errors.Wrapf(err, "while building asynchronous operation URL")
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
			log.C(ctx).Info("Getting asynchronous operation status for object")
			opResp, err := sendRequestWithRetry(ctx, caller, opReq)
			if err != nil {
				return "", err
			}
			defer closeResponseBody(ctx, opResp)

			opBody, err := ioutil.ReadAll(opResp.Body)
			if err != nil {
				return "", errors.Errorf("failed to read operation response body from asynchronous request: %v", err)
			}

			if opResp.StatusCode != http.StatusOK {
				return "", errors.Errorf("failed to get asynchronous object operation status. Received status: %d and body: %s", opResp.StatusCode, opBody)
			}

			var opStatus types.OperationStatus
			err = json.Unmarshal(opBody, &opStatus)
			if err != nil {
				return "", errors.Errorf("failed to unmarshal object operation status: %v", err)
			}

			if opStatus.State == types.OperationStateInProgress {
				log.C(ctx).Infof("The asynchronous object operation state is still: %q", types.OperationStateInProgress)
				continue
			}

			if opStatus.State != types.OperationStateSucceeded {
				return "", errors.Errorf("the asynchronous object operation finished with state: %q. Errors: %v", opStatus.State, opStatus.Errors)
			}

			log.C(ctx).Infof("The asynchronous operation status for object finished with state: %q", opStatus.State)

			if isCreateRequest { // async creation
				return opStatus.ResourceID, nil
			} else {
				return "", nil // to be able to reuse the async function, in case of deletion we return empty object ID because it is not needed in the delete case
			}

		case <-timeout:
			return "", errors.New("timeout waiting for asynchronous operation status to finish")
		}
	}
}

func (c *client) RetrieveMultipleResources(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) ([]string, error) {
	strURL, err := buildURL(c.cfg.RegionToInstanceConfig[region].SMURL, resources.GetURLPath(), SubaccountKey, subaccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "while building %s URL", resources.GetType())
	}

	body, err := c.executeSyncRequest(ctx, strURL, region)
	if err != nil {
		return nil, errors.Wrapf(err, "while executing request for listing %s for subaccount with ID: %q", resources.GetType(), subaccountID)
	}
	log.C(ctx).Infof("Successfully listed %s for subaccount: %q", resources.GetType(), subaccountID)

	err = json.Unmarshal(body, &resources)
	if err != nil {
		return nil, errors.Errorf("failed to unmarshal %s: %v", resources.GetType(), err)
	}

	resourceIDs, err := resourceMatchParams.MatchMultiple(resources)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("%d %s are found", len(resourceIDs), resources.GetType())

	return resourceIDs, nil
}

func sendRequestWithRetry(ctx context.Context, caller ExternalSvcCaller, req *http.Request) (*http.Response, error) {
	var response *http.Response
	err := retry.Do(func() error {
		res, err := caller.Call(req)
		if err != nil {
			return errors.Wrap(err, "failed to execute HTTP request")
		}

		if err == nil && res.StatusCode < http.StatusInternalServerError {
			response = res
			return nil
		}

		defer closeResponseBody(ctx, res)
		body, err := io.ReadAll(res.Body)

		if err != nil {
			return errors.Wrap(err, "failed to read response body")
		}
		return errors.Errorf("request failed with status code %d, error message: %v", res.StatusCode, string(body))
	},
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "connection refused") ||
				strings.Contains(err.Error(), "connection reset by peer")
		}))

	return response, err
}
