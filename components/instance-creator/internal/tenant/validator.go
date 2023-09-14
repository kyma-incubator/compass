package tenant

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/config"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"net/http"
)

// ClientIDFromCertificateHeader contains the name of the header containing the client id from the certificate
const ClientIDFromCertificateHeader = "Client-Id-From-Certificate"

// Middleware authorizes only requests made from CMP - checks if the tenant from ID token matches the provider subaccount of CMP
type Middleware struct {
	config      config.TenantInfo
	client      *http.Client
	certSubject string
}

// NewMiddleware provides new Middleware
func NewMiddleware(ctx context.Context, cfg config.TenantInfo) (Middleware, error) {
	middleware := Middleware{
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
func (m *Middleware) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			clientID := r.Header.Get(ClientIDFromCertificateHeader)
			if clientID == "" {
				log.C(ctx).Errorf("Failed to find client ID from header")
				apperrors.WriteAppError(ctx, w, errors.New("Tenant not found in request"), http.StatusBadRequest)
			}
			orgUnit := fmt.Sprintf("OU=%s", clientID)

			if !strings.Contains(m.certSubject, orgUnit) {
				log.C(ctx).Errorf("Tenant %s is not authorized", clientID)
				apperrors.WriteAppError(ctx, w, errors.Errorf("Tenant %s is not authorized", clientID), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *Middleware) getTenant(ctx context.Context) error {
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
