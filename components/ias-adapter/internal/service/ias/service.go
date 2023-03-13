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
		if err := s.updateApplication(ctx, iasHost, data.ConsumerApplication); err != nil {
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

	application := types.Application{}
	if err := json.NewDecoder(resp.Body).Decode(&application); err != nil {
		return types.Application{}, err
	}

	return application, nil
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

func (s Service) updateApplication(ctx context.Context, iasHost string, application types.Application) error {
	log := logger.FromContext(ctx)

	timeoutCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()

	applicationBytes, err := json.Marshal(application)
	if err != nil {
		return errors.Newf("failed to marshal body: %w", err)
	}

	url := buildPatchApplicationURL(iasHost, application.ID)
	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPatch, url, bytes.NewBuffer(applicationBytes))
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
			log.Warn().Msgf("failed to read response body for application with ID '%s': %s", application.ID, err)
		}
		return errors.Newf("failed to update ACL of application with ID '%s', status '%d', body '%s'",
			application.ID, resp.StatusCode, respBytes)
	}

	return nil
}

func buildGetApplicationURL(host, clientID string) string {
	return url.QueryEscape(fmt.Sprintf("%s%s/?filter=clientId eq %s", host, applicationsPath, clientID))
}

func buildPatchApplicationURL(host, applicationID string) string {
	return url.QueryEscape(fmt.Sprintf("%s%s/%s", host, applicationsPath, applicationID))
}
