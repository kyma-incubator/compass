package ias

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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
			removeConsumedAPI(&consumedAPIs, consumedAPI)
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

func (s Service) GetApplication(ctx context.Context, iasHost, clientID, appTenantID string) (types.Application, error) {
	log := logger.FromContext(ctx)

	url := buildGetApplicationURL(iasHost, clientID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.Application{}, errors.Newf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return types.Application{}, errors.Newf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().Msgf("failed to read GET application response body: %s", err)
		}
		return types.Application{}, errors.Newf("failed to get application, status '%d', body '%s'", resp.StatusCode, respBytes)
	}

	applications := types.Applications{}
	if err := json.NewDecoder(resp.Body).Decode(&applications); err != nil {
		return types.Application{}, err
	}

	if len(applications.Applications) == 0 {
		return types.Application{}, errors.Newf("no applications found with clientID '%s'", clientID)
	}
	if len(applications.Applications) == 1 {
		return applications.Applications[0], nil // TODO do we leave this?
	}

	return filterByAppTenantID(applications.Applications, clientID, appTenantID)
}

func filterByAppTenantID(applications []types.Application, clientID, appTenantID string) (types.Application, error) {
	for _, application := range applications {
		if application.Authentication.SAPManagedAttributes.AppTenantId == appTenantID {
			return application, nil
		}
	}
	return types.Application{}, errors.Newf(
		"application with clientID '%s' and appTenantID '%s' not found", clientID, appTenantID)
}

func addConsumedAPI(consumedAPIs *[]types.ApplicationConsumedAPI, consumedAPI types.ApplicationConsumedAPI) {
	for _, api := range *consumedAPIs {
		if api.APIName == consumedAPI.APIName {
			return
		}
	}
	*consumedAPIs = append(*consumedAPIs, consumedAPI)
}

func removeConsumedAPI(consumedAPIs *[]types.ApplicationConsumedAPI, apiName string) {
	found := false
	i := -1
	for i = range *consumedAPIs {
		if (*consumedAPIs)[i].APIName == apiName {
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

func buildGetApplicationURL(host, clientID string) string {
	escapedFilter := url.QueryEscape(fmt.Sprintf("clientId eq %s", clientID))
	return fmt.Sprintf("%s%s/?filter=%s", host, applicationsPath, escapedFilter)
}

func buildPatchApplicationURL(host, applicationID string) string {
	return fmt.Sprintf("%s%s/%s", host, applicationsPath, applicationID)
}
