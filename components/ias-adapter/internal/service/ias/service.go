package ias

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

const applicationsPath = "/Applications/v1"

type Service struct {
	cfg    config.IAS
	client *http.Client
}

func NewService(cfg config.IAS, client *http.Client) Service {
	return Service{
		cfg:    cfg,
		client: client,
	}
}

type UpdateData struct {
	Operation             types.Operation
	TenantMapping         types.TenantMapping
	ConsumerApplication   types.Application
	ProviderApplicationID string
}

func (s Service) UpdateApplicationConsumedAPIs(ctx context.Context, data UpdateData) error {
	consumerTenant := data.TenantMapping.AssignedTenants[0]
	consumedAPIs := data.ConsumerApplication.Authentication.ConsumedAPIs
	consumedAPIsLen := len(consumedAPIs)

	switch data.Operation {
	case types.OperationAssign:
		for _, consumedAPI := range consumerTenant.Configuration.ConsumedAPIs {
			addConsumedAPI(&consumedAPIs, types.ApplicationConsumedAPI{
				Name:    consumedAPI,
				APIName: consumedAPI,
				AppID:   data.ProviderApplicationID,
			})
		}
	case types.OperationUnassign:
		for _, consumedAPI := range consumerTenant.Configuration.ConsumedAPIs {
			removeConsumedAPI(&consumedAPIs, types.ApplicationConsumedAPI{
				Name:    consumedAPI,
				APIName: consumedAPI,
				AppID:   data.ProviderApplicationID,
			})
		}
	}

	if consumedAPIsLen != len(consumedAPIs) {
		iasHost := data.TenantMapping.ReceiverTenant.ApplicationURL
		if err := s.updateApplication(ctx, iasHost, data.ConsumerApplication.ID, consumedAPIs); err != nil {
			return errors.Newf("failed to update IAS application '%s' with UCL ID '%s': %w", data.ConsumerApplication.ID, consumerTenant.UCLApplicationID, err)
		}
	}

	return nil
}

func (s Service) GetApplicationByClientID(ctx context.Context, iasHost, clientID, appTenantID string) (types.Application, error) {
	filterByClientIDQuery := fmt.Sprintf("clientId eq %s", clientID)
	apps, err := s.getApplication(ctx, iasHost, filterByClientIDQuery)
	if err != nil {
		return types.Application{}, errors.Newf("could not get IAS application with client ID '%s': %w", clientID, err)
	}

	return filterByAppTenantID(apps, clientID, appTenantID)
}

func (s Service) GetApplicationByName(ctx context.Context, iasHost, name string) (types.Application, error) {
	filterByNameQuery := fmt.Sprintf("name eq %s", name)
	apps, err := s.getApplication(ctx, iasHost, filterByNameQuery)
	if err != nil {
		return types.Application{}, errors.Newf("could not get IAS application with name '%s': %w", name, err)
	}
	if len(apps) > 1 {
		return types.Application{}, errors.Newf("found %d IAS applications with name '%s'", len(apps), name)
	}

	return apps[0], nil
}

func (s Service) getApplication(ctx context.Context, iasHost, filterQuery string) ([]types.Application, error) {
	log := logger.FromContext(ctx)

	url := buildGetApplicationURL(iasHost, filterQuery)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []types.Application{}, errors.Newf("failed to create request: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return []types.Application{}, errors.Newf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return []types.Application{}, errors.Newf("no applications found with filter '%s': %w", filterQuery, errors.IASApplicationNotFound)
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().Msgf("Failed to read GET application response body: %s", err)
		}
		return []types.Application{}, errors.Newf("failed to get application, status '%d', body '%s'", resp.StatusCode, respBytes)
	}

	applications := types.Applications{}
	if err := json.NewDecoder(resp.Body).Decode(&applications); err != nil {
		return []types.Application{}, err
	}

	if len(applications.Applications) == 0 {
		return []types.Application{}, errors.Newf("no applications found with filter '%s': %w", filterQuery, errors.IASApplicationNotFound)
	}

	return applications.Applications, nil
}

func (s Service) CreateApplication(ctx context.Context, iasHost string, app *types.Application) (string, error) {
	log := logger.FromContext(ctx)
	url := buildCreateApplicationURL(iasHost)
	appBytes, err := json.Marshal(app)
	if err != nil {
		return "", errors.Newf("failed to marshal request body: %w", err)
	}
	log.Info().Msgf("Creating application with body: %s", appBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(appBytes))
	if err != nil {
		return "", errors.Newf("failed to create request: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", errors.Newf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().Msgf("Failed to read create application response body: %s", err)
		}
		return "", errors.Newf("failed to create application, status '%d', body '%s'", resp.StatusCode, respBytes)
	}

	appLocation := resp.Header.Get("Location")
	if appLocation == "" {
		return "", errors.Newf("created application location is empty")
	}
	appId := strings.TrimPrefix(appLocation, fmt.Sprintf("%s/", applicationsPath))
	log.Info().Msgf("IAS application with id '%s' created successfully", appId)

	return appId, nil
}

func filterByAppTenantID(applications []types.Application, clientID, appTenantID string) (types.Application, error) {
	for _, application := range applications {
		if application.Authentication.SAPManagedAttributes.AppTenantID == appTenantID ||
			application.Authentication.SAPManagedAttributes.SAPZoneID == appTenantID {
			return application, nil
		}
	}
	return types.Application{}, errors.Newf(
		"application with clientID '%s' and appTenantID '%s' not found: %w", clientID, appTenantID, errors.IASApplicationNotFound)
}

func addConsumedAPI(consumedAPIs *[]types.ApplicationConsumedAPI, consumedAPI types.ApplicationConsumedAPI) {
	for _, api := range *consumedAPIs {
		if api.APIName == consumedAPI.APIName && api.AppID == consumedAPI.AppID {
			return
		}
	}
	*consumedAPIs = append(*consumedAPIs, consumedAPI)
}

func removeConsumedAPI(consumedAPIs *[]types.ApplicationConsumedAPI, consumedAPI types.ApplicationConsumedAPI) {
	found := false
	i := -1
	for i = range *consumedAPIs {
		existingAPI := (*consumedAPIs)[i]
		if existingAPI.APIName == consumedAPI.APIName && existingAPI.AppID == consumedAPI.AppID {
			found = true
			break
		}
	}
	if !found {
		return
	}
	(*consumedAPIs)[i] = (*consumedAPIs)[len(*consumedAPIs)-1]
	*consumedAPIs = (*consumedAPIs)[:len(*consumedAPIs)-1]
}

func (s Service) updateApplication(ctx context.Context, iasHost, applicationID string,
	consumedAPIs []types.ApplicationConsumedAPI) error {

	log := logger.FromContext(ctx)

	appUpdate := types.ApplicationUpdate{
		Operations: []types.ApplicationUpdateOperation{
			{
				Operation: types.ReplaceOp,
				Path:      types.ConsumedAPIsPath,
				Value:     consumedAPIs,
			},
		},
	}
	appUpdateBytes, err := json.Marshal(appUpdate)
	if err != nil {
		return errors.Newf("failed to marshal body: %w", err)
	}
	log.Info().Msgf("Executing patch with body: %s", appUpdateBytes)
	url := buildPatchApplicationURL(iasHost, applicationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(appUpdateBytes))
	if err != nil {
		return errors.Newf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Newf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().Msgf("Failed to read response body for application with ID '%s': %s", applicationID, err)
		}
		return errors.Newf("request failed: status '%d', body '%s'", resp.StatusCode, respBytes)
	}

	return nil
}

func buildGetApplicationURL(host, filterQuery string) string {
	escapedFilter := url.QueryEscape(filterQuery)
	return fmt.Sprintf("%s%s/?filter=%s", host, applicationsPath, escapedFilter)
}

func buildCreateApplicationURL(host string) string {
	return fmt.Sprintf("%s%s/", host, applicationsPath)
}

func buildPatchApplicationURL(host, applicationID string) string {
	return fmt.Sprintf("%s%s/%s", host, applicationsPath, applicationID)
}
