package tenantvalidator

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/config"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"net/http"
)

// TenantValidationMiddleware authorizes only requests made from CMP - checks if the tenant from ID token matches the provider subaccount of CMP
type TenantValidationMiddleware struct {
	config      config.TenantInfo
	client      *http.Client
	certSubject string
}

// NewTenantValidationMiddleware provides new TenantValidationMiddleware
func NewTenantValidationMiddleware(ctx context.Context, cfg config.TenantInfo) (TenantValidationMiddleware, error) {
	middleware := TenantValidationMiddleware{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.InsecureSkipVerify,
				},
			},
		},
	}
	if err := middleware.getTenant(ctx); err != nil {
		return middleware, errors.Wrap(err, "failed to get auth tenant")
	}
	return middleware, nil
}

// Handler performs a tenant validation by comparing the tenant from the ID token with CMPs provider subaccount
func (m *TenantValidationMiddleware) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			tenant, err := authmiddleware.LoadExternalTenantFromContext(ctx)
			if err != nil {
				log.C(ctx).Error("Failed to find tenant in context")
				apperrors.WriteAppError(ctx, w, err, http.StatusBadRequest)
				return
			}
			orgUnit := fmt.Sprintf("OU=%s", tenant)

			if !strings.Contains(m.certSubject, orgUnit) {
				log.C(ctx).Errorf("Tenant %s is not authorized", tenant)
				apperrors.WriteAppError(ctx, w, errors.Errorf("Tenant %s is not authorized", tenant), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *TenantValidationMiddleware) getTenant(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.config.Endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute GET request")
	}
	defer resp.Body.Close()

	tenantInfo := TenantInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&tenantInfo); err != nil {
		return errors.Wrap(err, "failed to decode response")
	}

	m.certSubject = tenantInfo.CertSubject
	return nil
}
