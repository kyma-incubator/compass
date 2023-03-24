package middlewares

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

type AuthMiddleware struct {
	Config      config.TenantInfo
	client      *http.Client
	certSubject string
}

func NewAuthMiddleware(ctx context.Context, cfg config.TenantInfo) (AuthMiddleware, error) {
	middleware := AuthMiddleware{
		Config: cfg,
	}
	httpClient, err := newInfoClient(cfg)
	if err != nil {
		return middleware, errors.Newf("failed to create info http client: %w", err)
	}
	middleware.client = httpClient
	if err := middleware.getTenant(ctx); err != nil {
		return middleware, errors.Newf("failed to get auth tenant: %w", err)
	}
	return middleware, nil
}

func newInfoClient(cfg config.TenantInfo) (*http.Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
	}
	if cfg.InsecureSkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.InsecureSkipVerify,
			},
		}
		return httpClient, nil
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM([]byte(cfg.RootCA)); !ok {
		return nil, errors.New("failed to append info endpoint root CA")
	}
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	return httpClient, nil
}

func (m *AuthMiddleware) Auth(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	tenant, exists := ctx.Get(tenantCtxKey)
	if !exists {
		log.Error().Msg("Failed to find tenant in context")
		internal.RespondWithError(ctx, http.StatusInternalServerError, errors.New(""))
		return
	}
	orgUnit := fmt.Sprintf("OU=%s", tenant)
	if !strings.Contains(m.certSubject, orgUnit) {
		log.Error().Msgf("Tenant %s is not authorized", tenant)
		internal.RespondWithError(ctx, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	ctx.Next()
}

func (m *AuthMiddleware) getTenant(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.Config.Endpoint, nil)
	if err != nil {
		return errors.Newf("failed to create request: %w", err)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return errors.Newf("failed to execute GET request: %w", err)
	}
	defer resp.Body.Close()

	tenantInfo := types.TenantInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&tenantInfo); err != nil {
		return errors.Newf("failed to decode response: %w", err)
	}

	m.certSubject = tenantInfo.CertSubject
	return nil
}
