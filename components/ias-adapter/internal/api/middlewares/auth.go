package middlewares

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

type AuthMiddleware struct {
	Config config.TenantInfo
	tenant string
}

func NewAuthMiddleware(ctx context.Context, cfg config.TenantInfo) (AuthMiddleware, error) {
	middleware := AuthMiddleware{Config: cfg}
	if err := middleware.getTenant(ctx); err != nil {
		return middleware, errors.Newf("failed to get auth tenant: %w", err)
	}
	return middleware, nil
}

func (m *AuthMiddleware) Auth(ctx *gin.Context) {
	log := logger.FromContext(ctx)
	log.Info().Msgf("Header '%s': %s", "X-Forwarded-Client-Cert", ctx.Request.Header.Get("X-Forwarded-Client-Cert"))

	ctx.Next()
}

func (m *AuthMiddleware) getTenant(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), m.Config.RequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, m.Config.Endpoint, nil)
	if err != nil {
		return errors.Newf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Newf("failed to execute GET request: %w", err)
	}
	defer resp.Body.Close()

	tenantInfo := types.TenantInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&tenantInfo); err != nil {
		return errors.Newf("failed to decode response: %w", err)
	}

	m.tenant = tenantInfo.Tenant()
	return nil
}
