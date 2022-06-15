package appmetadatavalidation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gqlgen "github.com/99designs/gqlgen/graphql"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// LabelService is responsible to service-related label operations
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

// TenantService is responsible to service-related tenant operations
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// Handler is an object with dependencies
type Handler struct {
	transact  persistence.Transactioner
	tenantSvc TenantService
	labelSvc  LabelService
}

// NewHandler is a constructor for Handler object
func NewHandler(transact persistence.Transactioner, tenantSvc TenantService, labelSvc LabelService) *Handler {
	return &Handler{
		transact:  transact,
		tenantSvc: tenantSvc,
		labelSvc:  labelSvc,
	}
}

// Handler is a middleware that checks if the flow is oathkeeper.CertificateFlow and consumer type is consumer.ExternalCertificate.
// If it's not - continue with next handler. If it is, get the consumer tenant's Region label and the Region label of the tenant header.
// They have to match in order to continue with the next handler, otherwise fail the request
func (h *Handler) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.C(ctx)

			consumerInfo, err := consumer.LoadFromContext(ctx)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while fetching consumer info from context: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			if !h.isExternalCertificateFlow(consumerInfo) {
				logger.Infof("Flow is not external certificate. Continue with next handler.")
				next.ServeHTTP(w, r)
				return
			}

			tenantHeader := r.Header.Get("tenant")
			if len(tenantHeader) == 0 {
				logger.Infof("Missing tenant header. Continue with next handler.")
				next.ServeHTTP(w, r)
				return
			}
			logger.Infof("Flow is external certificate")

			tx, err := h.transact.Begin()
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while opening transaction: %v", err)
				appErr := apperrors.InternalErrorFrom(err, "while opening db transaction")
				writeAppError(ctx, w, appErr)
				return
			}

			defer h.transact.RollbackUnlessCommitted(ctx, tx)

			ctx = persistence.SaveToContext(ctx, tx)

			consumerRegionLabel, err := h.getTenantRegionLabelValue(ctx, consumerInfo.ConsumerID)
			if err != nil {
				appErr := apperrors.InternalErrorFrom(err, "while fetching consumer tenant %q region label", consumerInfo.ConsumerID)
				writeAppError(ctx, w, appErr)
				return
			}

			headerTenantRegionLabel, err := h.getTenantRegionLabelValue(ctx, tenantHeader)
			if err != nil {
				appErr := apperrors.InternalErrorFrom(err, "while fetching tenant header %q region label", tenantHeader)
				writeAppError(ctx, w, appErr)
				return
			}

			if consumerRegionLabel != headerTenantRegionLabel {
				err = errors.New(fmt.Sprintf("region labels mismatch: %q and %q for tenants %q and %q", consumerRegionLabel, headerTenantRegionLabel, consumerInfo.ConsumerID, tenantHeader))
				logger.WithError(err).Errorf("Regions for external tenants %q and %q do not match: %q and %q", consumerInfo.ConsumerID, tenantHeader, consumerRegionLabel, headerTenantRegionLabel)
				appErr := apperrors.InternalErrorFrom(err, "region labels for tenants do not match")
				writeAppError(ctx, w, appErr)
				return
			}

			if err := tx.Commit(); err != nil {
				logger.WithError(err).Errorf("An error has occurred while committing transaction: %v", err)
				appErr := apperrors.InternalErrorFrom(err, "error has occurred while committing transaction")
				writeAppError(ctx, w, appErr)
				return
			}

			logger.Infof("Regions for tenants with external ID %q and %q matched. Continuing with request", consumerInfo.ConsumerID, tenantHeader)

			next.ServeHTTP(w, r)
		})
	}
}

func (h *Handler) getTenantRegionLabelValue(ctx context.Context, tenantID string) (interface{}, error) {
	logger := log.C(ctx)

	tenantModel, err := h.tenantSvc.GetTenantByExternalID(ctx, tenantID)
	if err != nil {
		logger.WithError(err).Errorf("An error has occurred while fetching tenant by external ID %q: %v", tenantID, err)
		return nil, errors.Wrapf(err, "while fetching tenant by external ID %q", tenantID)
	}

	tenantRegionLabel, err := h.labelSvc.GetByKey(ctx, tenantModel.ID, model.TenantLabelableObject, tenantModel.ID, tnt.RegionLabelKey)
	if err != nil {
		logger.WithError(err).Errorf("An error has occurred while fetching %q label for tenant ID %q: %v", tnt.RegionLabelKey, tenantID, err)
		return nil, errors.Wrapf(err, "while fetching %q label tenant by external ID %q", tnt.RegionLabelKey, tenantID)
	}

	return tenantRegionLabel.Value, nil
}

func (h *Handler) isExternalCertificateFlow(consumerInfo consumer.Consumer) bool {
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

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}
