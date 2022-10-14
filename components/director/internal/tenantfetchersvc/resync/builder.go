package resync

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	graphqlclient "github.com/kyma-incubator/compass/components/director/pkg/graphql_client"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

type synchronizerBuilder struct {
	jobConfig                JobConfig
	featuresConfig           features.Config
	transact                 persistence.Transactioner
	directorClient           *graphqlclient.Director
	aggregationFailurePusher AggregationFailurePusher
}

// NewSynchronizerBuilder returns an entity that will use the provided configuration to create a tenant synchronizer.
func NewSynchronizerBuilder(jobConfig JobConfig, featuresConfig features.Config, transact persistence.Transactioner, directorClient *graphqlclient.Director, aggregationFailurePusher AggregationFailurePusher) *synchronizerBuilder {
	return &synchronizerBuilder{
		jobConfig:                jobConfig,
		featuresConfig:           featuresConfig,
		transact:                 transact,
		directorClient:           directorClient,
		aggregationFailurePusher: aggregationFailurePusher,
	}
}

// Build returns a tenants synchronizer created with the initially provided configuration of the builder.
func (b *synchronizerBuilder) Build(ctx context.Context) (*TenantsSynchronizer, error) {
	kubeClient, err := NewKubernetesClient(ctx, b.jobConfig.KubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Kubernetes client for job %s", b.jobConfig.JobName)
	}

	universalEventAPIClient, additionalRegionalEventAPIClients, err := b.eventAPIClients()
	if err != nil {
		return nil, err
	}

	tenantSvc, tenantConverter, runtimeSvc, labelRepo := domainServices(b.featuresConfig)
	tenantManager, err := NewTenantsManager(b.jobConfig, b.directorClient, universalEventAPIClient, additionalRegionalEventAPIClients, tenantConverter)
	if err != nil {
		return nil, err
	}

	var mover TenantMover
	if b.jobConfig.TenantType == tenantpkg.Subaccount {
		mover = NewSubaccountsMover(b.jobConfig, b.transact, b.directorClient, universalEventAPIClient, tenantConverter, tenantSvc, runtimeSvc, labelRepo)
	} else {
		mover = newNoOpsMover()
	}

	ts := NewTenantSynchronizer(b.jobConfig, b.transact, tenantSvc, tenantManager, mover, tenantManager, kubeClient, b.aggregationFailurePusher)
	return ts, nil
}

func (b *synchronizerBuilder) eventAPIClients() (EventAPIClient, map[string]EventAPIClient, error) {
	clientCfg := ClientConfig{
		TenantProvider:      b.jobConfig.TenantProvider,
		APIConfig:           b.jobConfig.APIConfig.APIEndpointsConfig,
		FieldMapping:        b.jobConfig.APIConfig.TenantFieldMapping,
		MovedSAFieldMapping: b.jobConfig.APIConfig.MovedSubaccountsFieldMapping,
	}
	eventAPIClient, err := NewClient(b.jobConfig.APIConfig.OAuthConfig, b.jobConfig.APIConfig.AuthMode, clientCfg, b.jobConfig.APIConfig.ClientTimeout)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating a Event API client")
	}

	regionalEventAPIClients := map[string]EventAPIClient{}
	for _, config := range b.jobConfig.RegionalAPIConfigs {
		clientConfig := ClientConfig{
			TenantProvider:      b.jobConfig.TenantProvider,
			APIConfig:           config.APIEndpointsConfig,
			FieldMapping:        config.TenantFieldMapping,
			MovedSAFieldMapping: config.MovedSubaccountsFieldMapping,
		}
		eventAPIClient, err := NewClient(config.OAuthConfig, config.AuthMode, clientConfig, config.ClientTimeout)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while creating a regional Event API client for region %s", config.RegionName)
		}
		regionalEventAPIClients[config.RegionName] = eventAPIClient
	}

	return eventAPIClient, regionalEventAPIClients, nil
}

func domainServices(featuresConfig features.Config) (TenantStorageService, TenantConverter, RuntimeService, LabelRepo) {
	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	tenantStorageConverter := tenant.NewConverter()
	labelConverter := label.NewConverter()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	scenarioAssignConverter := scenarioassignment.NewConverter()
	runtimeContextConverter := runtimectx.NewConverter()
	tenantConverter := tenant.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter()

	webhookRepo := webhook.NewRepository(webhookConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelRepo := label.NewRepository(labelConverter)
	tenantStorageRepo := tenant.NewRepository(tenantStorageConverter)
	applicationRepo := application.NewRepository(appConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(scenarioAssignConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepo, labelSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantStorageRepo, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, labelDefSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)

	notificationSvc := formation.NewNotificationService(applicationRepo, nil, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookConverter, nil)
	formationSvc := formation.NewService(labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tenantSvc, runtimeRepo, runtimeContextRepo, notificationSvc, featuresConfig.RuntimeTypeLabelKey, featuresConfig.ApplicationTypeLabelKey)
	runtimeContextSvc := runtimectx.NewService(runtimeContextRepo, labelRepo, runtimeRepo, labelSvc, formationSvc, tenantSvc, uidSvc)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, labelSvc, uidSvc, formationSvc, tenantStorageSvc, webhookSvc, runtimeContextSvc, featuresConfig.ProtectedLabelPattern, featuresConfig.ImmutableLabelPattern, featuresConfig.RuntimeTypeLabelKey, featuresConfig.KymaRuntimeTypeLabelValue)

	return tenantSvc, tenantConverter, runtimeSvc, labelRepo
}
