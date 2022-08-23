package destinationfetchersvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	tenantIDKey        = "subaccountId"
	destQueryParameter = "dest"
)

// HandlerConfig destination handler configuration
type HandlerConfig struct {
	SyncDestinationsEndpoint      string `envconfig:"APP_DESTINATIONS_SYNC_ENDPOINT,default=/v1/syncDestinations"`
	DestinationsSensitiveEndpoint string `envconfig:"APP_DESTINATIONS_SENSITIVE_DATA_ENDPOINT,default=/v1/destinations"`
	UserContextHeader             string `envconfig:"APP_USER_CONTEXT_HEADER,default=user_context"`
}

type handler struct {
	destinationManager DestinationManager
	config             HandlerConfig
}

//go:generate mockery --name=DestinationManager --output=automock --outpkg=automock --case=underscore --disable-version-string
// DestinationManager missing godoc
type DestinationManager interface {
	SyncTenantDestinations(ctx context.Context, tenantID string) error
	FetchDestinationsSensitiveData(ctx context.Context, tenantID string, destinationNames []string) ([]byte, error)
}

// NewDestinationsHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewDestinationsHTTPHandler(destinationManager DestinationManager, config HandlerConfig) *handler {
	return &handler{
		destinationManager: destinationManager,
		config:             config,
	}
}

func (h *handler) SyncTenantDestinations(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	userContextHeader := request.Header.Get(h.config.UserContextHeader)
	tenantID, err := h.readTenantFromHeader(userContextHeader)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.destinationManager.SyncTenantDestinations(ctx, tenantID); err != nil {
		if apperrors.IsNotFoundError(err) {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(writer, fmt.Sprintf("Failed to sync destinations for tenant %s",
			tenantID), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getDestinationNames(namesRaw string) ([]string, error) {
	namesRawLength := len(namesRaw)
	if namesRawLength == 0 {
		return nil, fmt.Errorf("dest query parameter is missing")
	}

	if namesRaw[0] != '[' || namesRaw[namesRawLength-1] != ']' {
		return nil, fmt.Errorf("%s dest query parameter is invalid. Must start with '[' and end with ']'", namesRaw)
	}

	// Remove brackets from query
	namesRawWithoutBrackets := namesRaw[1 : namesRawLength-1]
	names := strings.Split(namesRawWithoutBrackets, ",")

	if sliceContainsEmptyString(names) {
		return nil, fmt.Errorf("name parameter contains empty element")
	}

	for idx, name := range names {
		names[idx] = strings.Trim(name, " ")
	}

	return names, nil
}

func (h *handler) FetchDestinationsSensitiveData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	userContextHeader := request.Header.Get(h.config.UserContextHeader)
	tenantID, err := h.readTenantFromHeader(userContextHeader)
	if err != nil {
		log.C(ctx).Errorf("Failed to read userContext header with error: %s", err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	namesRaw := request.URL.Query().Get(destQueryParameter)
	names, err := getDestinationNames(namesRaw)
	if err != nil {
		log.C(ctx).Errorf("Failed to fetch sensitive data for destinations %s: %v", namesRaw, err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	json, err := h.destinationManager.FetchDestinationsSensitiveData(ctx, tenantID, names)

	if err != nil {
		log.C(ctx).Errorf("Failed to fetch sensitive data for destinations %s and tenant %s: %v",
			namesRaw, tenantID, err)
		if apperrors.IsNotFoundError(err) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(writer, fmt.Sprintf("Failed to fetch sensitive data for destinations %s", namesRaw), http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(json); err != nil {
		log.C(ctx).WithError(err).Error("Could not write response")
	}
}

func sliceContainsEmptyString(s []string) bool {
	for _, e := range s {
		if strings.TrimSpace(e) == "" {
			return true
		}
	}

	return false
}

func (h *handler) readTenantFromHeader(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("%s header is missing", h.config.UserContextHeader)
	}

	var headerMap map[string]string
	if err := json.Unmarshal([]byte(header), &headerMap); err != nil {
		return "", fmt.Errorf("failed to parse %s header", h.config.UserContextHeader)
	}

	tenantID, ok := headerMap[tenantIDKey]
	if !ok {
		return "", fmt.Errorf("%s not found in %s header", tenantIDKey, h.config.UserContextHeader)
	}

	return tenantID, nil
}
