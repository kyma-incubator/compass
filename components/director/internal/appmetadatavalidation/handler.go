package appmetadatavalidation

import (
	"context"
	"encoding/json"
	gqlgen "github.com/99designs/gqlgen/graphql"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
)

// LabelService missing godoc
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

// TenantService missing godoc
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// Handler missing godoc
type Handler struct {
	transact  persistence.Transactioner
	tenantSvc TenantService
	labelSvc  LabelService
}

// NewHandler missing godoc
func NewHandler(transact persistence.Transactioner, tenantSvc TenantService, labelSvc LabelService) *Handler {
	return &Handler{
		transact:  transact,
		tenantSvc: tenantSvc,
		labelSvc:  labelSvc,
	}
}

// Handler missing godoc
func (h *Handler) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.C(ctx)

			consumerInfo, err := consumer.LoadFromContext(ctx)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while fetching consumer info from context: %v", err)
				appErr := apperrors.InternalErrorFrom(err, "while fetching consumer info")
				writeAppError(ctx, w, appErr)
				return
			}

			if !isExternalCertificateFlow(consumerInfo) {
				next.ServeHTTP(w, r)
				return
			}

			// TODO check if empty
			headerTenant := r.Header.Get("tenant")

			consumerTenantModel, err := h.tenantSvc.GetTenantByExternalID(ctx, consumerInfo.ConsumerID)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while fetching consumer tenant by external ID %q: %v", consumerInfo.ConsumerID, err)
				appErr := apperrors.InternalErrorFrom(err, "while fetching consumer tenant by ID")
				writeAppError(ctx, w, appErr)
				return
			}

			consumerRegionLabel, err := h.labelSvc.GetByKey(ctx, consumerTenantModel.ID, model.TenantLabelableObject, consumerTenantModel.ID, tnt.RegionLabelKey)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while fetching %q label for tenant ID %q: %v", tnt.RegionLabelKey, consumerInfo.ConsumerID, err)
				appErr := apperrors.InternalErrorFrom(err, "while fetching label")
				writeAppError(ctx, w, appErr)
				return
			}

			headerTenantModel, err := h.tenantSvc.GetTenantByExternalID(ctx, headerTenant)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while fetching header tenant by external ID %q: %v", headerTenantModel.ID, err)
				appErr := apperrors.InternalErrorFrom(err, "while fetching header tenant by ID")
				writeAppError(ctx, w, appErr)
				return
			}

			headerTenantRegionLabel, err := h.labelSvc.GetByKey(ctx, headerTenantModel.ID, model.TenantLabelableObject, headerTenantModel.ID, tnt.RegionLabelKey)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while fetching %q label for tenant ID %q: %v", tnt.RegionLabelKey, headerTenantModel.ID, err)
				appErr := apperrors.InternalErrorFrom(err, "while fetching label")
				writeAppError(ctx, w, appErr)
				return
			}

			if consumerRegionLabel.Value != headerTenantRegionLabel.Value {
				logger.WithError(err).Errorf("Regions do not match: %q and %q", consumerRegionLabel.Value, headerTenantRegionLabel.Value)

				appErr := apperrors.InternalErrorFrom(err, "while reading request body")
				writeAppError(ctx, w, appErr)
			}

			logger.Infof("Regions matched. Continuing with request")

			next.ServeHTTP(w, r)
		})
	}
}

func isExternalCertificateFlow(consumerInfo consumer.Consumer) bool {
	return consumerInfo.Flow.IsCertFlow() && consumerInfo.ConsumerType == consumer.ExternalCertificate
}

func writeAppError(ctx context.Context, w http.ResponseWriter, appErr error) {
	errCode := apperrors.ErrorCode(appErr)
	if errCode == apperrors.UnknownError || errCode == apperrors.InternalError {
		errCode = apperrors.InternalError
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	resp := gqlgen.Response{Errors: []*gqlerror.Error{{
		Message:    appErr.Error(),
		Extensions: map[string]interface{}{"error_code": errCode, "error": errCode.String()}}}}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}
