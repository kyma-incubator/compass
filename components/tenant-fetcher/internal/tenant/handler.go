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
	HandlerEndpoint string `mapstructure:"handler_endpoint"`
	TenantPathParam string `mapstructure:"tenant_path_param"`

	TenantProviderTenantIdProperty string `mapstructure:"tenant_provider_tenant_id_property"`
	TenantProvider                 string `mapstructure:"tenant_provider"`

	JWKSSyncPeriod            time.Duration `mapstructure:"jwks_sync_period"`
	AllowJWTSigningNone       bool          `mapstructure:"allow_jwt_signing_none"`
	JwksEndpoints             string        `mapstructure:"jwks_endpoints"`
	IdentityZone              string        `mapstructure:"tenant_identity_zone"`
	SubscriptionCallbackScope string        `mapstructure:"subscription_callback_scope"`
}

func DefaultConfig() *Config {
	return &Config{
		HandlerEndpoint:                "/v1/callback/{tenantId}",
		TenantPathParam:                "tenantId",
		TenantProviderTenantIdProperty: "",
		TenantProvider:                 "",
		JWKSSyncPeriod:                 5 * time.Minute,
		AllowJWTSigningNone:            true,
		JwksEndpoints:                  "",
		IdentityZone:                   "",
		SubscriptionCallbackScope:      "",
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
