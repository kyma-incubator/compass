package destinationfetchersvc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// HandlerConfig destination handler configuration
type HandlerConfig struct {
	SyncDestinationsEndpoint      string `envconfig:"APP_DESTINATIONS_SYNC_ENDPOINT,default=/v1/syncDestinations"`
	DestinationsSensitiveEndpoint string `envconfig:"APP_DESTINATIONS_SENSITIVE_DATA_ENDPOINT,default=/v1/destinations"`
	DestinationsQueryParameter    string `envconfig:"APP_DESTINATIONS_SENSITIVE_DATA_QUERY_PARAM,default=name"`
}

type handler struct {
	destinationManager DestinationManager
	config             HandlerConfig
}

//go:generate mockery --name=DestinationManager --output=automock --outpkg=automock --case=underscore --disable-version-string
// DestinationManager missing godoc
type DestinationManager interface {
	IsTenantSubscribed(ctx context.Context, tenantID string) (bool, error)
	GetSubscribedTenantIDs(ctx context.Context) ([]string, error)
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

	tenantID, err := tenant.LoadFromContext(request.Context())
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to load tenant ID from request")
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	isTenantSubscribed, err := h.destinationManager.IsTenantSubscribed(ctx, tenantID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to validate tenant %q subscription", tenantID)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if !isTenantSubscribed {
		log.C(ctx).Infof("Tenant %q is not subscribed", tenantID)
		http.Error(writer, fmt.Sprintf("Tenant %q is not subscribed", tenantID), http.StatusInternalServerError)
		return
	}

	if err := h.destinationManager.SyncTenantDestinations(ctx, tenantID); err != nil {
		if apperrors.IsNotFoundError(err) {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		log.C(ctx).WithError(err).Errorf("Failed to sync destinations for tenant '%s'", tenantID)

		http.Error(writer, fmt.Sprintf("Failed to sync destinations for tenant %s",
			tenantID), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (h *handler) FetchDestinationsSensitiveData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	tenantID, err := tenant.LoadFromContext(request.Context())
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to fetch sensitive data for destinations")
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	isTenantSubscribed, err := h.destinationManager.IsTenantSubscribed(ctx, tenantID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to validate tenant %q subscription", tenantID)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if !isTenantSubscribed {
		log.C(ctx).Infof("Tenant %q is not subscribed", tenantID)
		http.Error(writer, fmt.Sprintf("Tenant %q is not subscribed", tenantID), http.StatusInternalServerError)
		return
	}

	names, ok := request.URL.Query()[h.config.DestinationsQueryParameter]
	if !ok {
		err := fmt.Errorf("missing query parameter '%s'", h.config.DestinationsQueryParameter)
		log.C(ctx).WithError(err).Error("While fetching destinations sensitive data")
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	json, err := h.destinationManager.FetchDestinationsSensitiveData(ctx, tenantID, names)

	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to fetch sensitive data for destinations %+v and tenant %s",
			names, tenantID)
		if apperrors.IsNotFoundError(err) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(writer, fmt.Sprintf("Failed to fetch sensitive data for destinations %+v", names), http.StatusInternalServerError)
		return
	}

	if _, err = writer.Write(json); err != nil {
		log.C(ctx).WithError(err).Error("Could not write response")
	}
}
