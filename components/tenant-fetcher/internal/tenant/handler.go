package tenant

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	auth "github.com/kyma-incubator/compass/components/tenant-fetcher/internal/authenticator"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/uuid"
)

type Config struct {
	HandlerEndpoint string `mapstructure:"HANDLER_ENDPOINT"`
	TenantPathParam string `mapstructure:"TENANT_PATH_PARAM"`

	TenantProviderTenantIdProperty string `mapstructure:"TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	TenantProvider                 string `mapstructure:"TENANT_PROVIDER"`

	JWKSSyncPeriod            time.Duration `mapstructure:"JWKS_SYNC_PERIOD"`
	AllowJWTSigningNone       bool          `mapstructure:"ALLOW_JWT_SIGNING_NONE,default=true"`
	JwksEndpoints             string        `mapstructure:"JWKS_ENDPOINTS"`
	IdentityZone              string        `mapstructure:"TENANT_IDENTITY_ZONE"`
	SubscriptionCallbackScope string        `mapstructure:"SUBSCRIPTION_CALLBACK_SCOPE"`
}

func DefaultConfig() *Config {
	return &Config{
		HandlerEndpoint:     "/v1/callback/{tenantId}",
		TenantPathParam:     "tenantId",
		JWKSSyncPeriod:      5 * time.Minute,
		AllowJWTSigningNone: true,
	}
}

const compassURL = "https://github.com/kyma-incubator/compass"

func RegisterHandler(ctx context.Context, router *mux.Router, cfg Config, authConfig []authenticator.Config, transact persistence.Transactioner) error {
	logger := log.C(ctx)

	var jwks []string

	if err := json.Unmarshal([]byte(cfg.JwksEndpoints), &jwks); err != nil {
		return apperrors.NewInternalError("unable to unmarshal jwks endpoints environment variable")
	}

	middleware := auth.New(
		jwks,
		cfg.IdentityZone,
		cfg.SubscriptionCallbackScope,
		extractTrustedIssuersScopePrefixes(authConfig),
		cfg.AllowJWTSigningNone,
	)

	router.Use(middleware.Handler())

	logger.Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
	periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(ctx context.Context) {
		err := middleware.SynchronizeJWKS(ctx)
		if err != nil {
			logger.WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
		}
	})
	go periodicExecutor.Run(ctx)

	uidSvc := uuid.NewService()
	converter := NewConverter()
	repo := NewRepository(converter)
	service := NewService(repo, transact, uidSvc, cfg)

	logger.Infof("Registering Tenant Onboarding endpoint on %s...", cfg.HandlerEndpoint)
	router.HandleFunc(cfg.HandlerEndpoint, service.Create).Methods(http.MethodPut)

	logger.Infof("Registering Tenant Decommissioning endpoint on %s...", cfg.HandlerEndpoint)
	router.HandleFunc(cfg.HandlerEndpoint, service.DeleteByExternalID).Methods(http.MethodDelete)

	return nil
}

func extractTrustedIssuersScopePrefixes(config []authenticator.Config) []string {
	var prefixes []string

	for _, authenticator := range config {
		if len(authenticator.TrustedIssuers) == 0 {
			continue
		}

		for _, trustedIssuers := range authenticator.TrustedIssuers {
			prefixes = append(prefixes, trustedIssuers.ScopePrefix)
		}
	}

	return prefixes
}
