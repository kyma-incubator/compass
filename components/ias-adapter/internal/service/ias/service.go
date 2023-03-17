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
	TenantMapping         types.TenantMapping
	ConsumerApplication   types.Application
	ProviderApplicationID string
}

func (s Service) UpdateApplicationConsumedAPIs(ctx context.Context, data UpdateData) error {
	oldConsumedAPIs := data.ConsumerApplication.Authentication.ConsumedAPIs
	var newConsumedAPIs []types.ApplicationConsumedAPI
	switch {
	case data.TenantMapping.AssignedTenants[0].Operation == types.OperationAssign:
		for _, consumedAPI := range data.TenantMapping.AssignedTenants[0].Configuration.ConsumedAPIs {
			newConsumedAPIs = addConsumedAPI(oldConsumedAPIs, types.ApplicationConsumedAPI{
				Name:    consumedAPI,
				APIName: consumedAPI,
				AppID:   data.ProviderApplicationID,
			})
		}
	case data.TenantMapping.AssignedTenants[0].Operation == types.OperationUnassign:
		for _, consumedAPI := range data.TenantMapping.AssignedTenants[0].Configuration.ConsumedAPIs {
			newConsumedAPIs = removeConsumedAPI(oldConsumedAPIs, consumedAPI)
		}
	}

	if len(oldConsumedAPIs) != len(newConsumedAPIs) {
		iasHost := data.TenantMapping.ReceiverTenant.ApplicationURL
		if err := s.updateApplication(ctx, iasHost, data.ConsumerApplication.ID, newConsumedAPIs); err != nil {
			return errors.Newf("failed to update application: %w", err)
		}
	}

	return nil
}

func (s Service) GetApplication(ctx context.Context, iasHost, clientID string) (types.Application, error) {
	log := logger.FromContext(ctx)

	timeoutCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()

	url := buildGetApplicationURL(iasHost, clientID)
	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, url, nil)
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
		return types.Application{}, errors.Newf("application with clientID '%s' not found", clientID)
	}
	if len(applications.Applications) > 1 {
		return types.Application{}, errors.Newf("found more than one application with clientID '%s'", clientID)
	}

	return applications.Applications[0], nil
}

func addConsumedAPI(consumedAPIs []types.ApplicationConsumedAPI, consumedAPI types.ApplicationConsumedAPI) []types.ApplicationConsumedAPI {
	for _, api := range consumedAPIs {
		if api.APIName == consumedAPI.APIName {
			return consumedAPIs
		}
	}
	return append(consumedAPIs, consumedAPI)
}

func removeConsumedAPI(consumedAPIs []types.ApplicationConsumedAPI, apiName string) []types.ApplicationConsumedAPI {
	found := false
	i := -1
	for i = range consumedAPIs {
		if consumedAPIs[i].APIName == apiName {
			found = true
			break
		}
	}
	if !found {
		return consumedAPIs
	}
	consumedAPIs[i] = consumedAPIs[len(consumedAPIs)-1]
	return consumedAPIs[:len(consumedAPIs)-1]
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
	timeoutCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPatch, url, bytes.NewBuffer(appUpdateBytes))
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
			log.Warn().Msgf("failed to read response body for application with ID '%s': %s", applicationID, err)
		}
		return errors.Newf("failed to update ACL of application with ID '%s', status '%d', body '%s'",
			applicationID, resp.StatusCode, respBytes)
	}

	return nil
}

func buildGetApplicationURL(host, clientID string) string {
	escapedFilter := url.QueryEscape(fmt.Sprintf("clientId eq %s", clientID))
	return fmt.Sprintf("%s%s?filter=%s", host, applicationsPath, escapedFilter)
}

func buildPatchApplicationURL(host, applicationID string) string {
	return fmt.Sprintf("%s%s/%s", host, applicationsPath, applicationID)
}
