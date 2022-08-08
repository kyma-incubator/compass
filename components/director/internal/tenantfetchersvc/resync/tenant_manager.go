package resync

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// EventAPIClient missing godoc
//go:generate mockery --name=EventAPIClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventAPIClient interface {
	FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (*EventsPage, error)
}

// TenantConverter expects tenant converter implementation
//go:generate mockery --name=TenantConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantConverter interface {
	MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput) []graphql.BusinessTenantMappingInput
	ToGraphQLInput(model.BusinessTenantMappingInput) graphql.BusinessTenantMappingInput
}

// DirectorGraphQLClient expects graphql implementation
//go:generate mockery --name=DirectorGraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorGraphQLClient interface {
	WriteTenants(context.Context, []graphql.BusinessTenantMappingInput) error
	DeleteTenants(ctx context.Context, tenants []graphql.BusinessTenantMappingInput) error
	UpdateTenant(ctx context.Context, id string, tenant graphql.BusinessTenantMappingInput) error
}

type supportedEvents struct {
	createdTenantEvent EventsType
	updatedTenantEvent EventsType
	deletedTenantEvent EventsType
}

type externalTenantsManager struct {
	config         EventsConfig
	tenantProvider string

	gqlClient      DirectorGraphQLClient
	eventAPIClient EventAPIClient

	tenantConverter TenantConverter
}

type TenantsManager struct {
	externalTenantsManager

	regionalClients     map[string]RegionalClient
	supportedEventTypes supportedEvents
}

func NewTenantsManager(jobConfig JobConfig, directorClient DirectorGraphQLClient, universalClient EventAPIClient, regionalDetails map[string]RegionalClient, tenantConverter TenantConverter) (*TenantsManager, error) {
	supportedEvents, err := supportedEventTypes(jobConfig.TenantType)
	if err != nil {
		return nil, err
	}

	tenantsManager := &TenantsManager{
		regionalClients:     regionalDetails,
		supportedEventTypes: supportedEvents,
		externalTenantsManager: externalTenantsManager{
			gqlClient:       directorClient,
			eventAPIClient:  universalClient,
			config:          jobConfig.EventsConfig,
			tenantConverter: tenantConverter,
			tenantProvider:  jobConfig.TenantProvider,
		},
	}
	return tenantsManager, nil
}

func (tm *TenantsManager) TenantsToCreate(ctx context.Context, region, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	fetchedTenants, err := tm.fetchTenantsForEventType(ctx, region, fromTimestamp, tm.supportedEventTypes.createdTenantEvent)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching created tenants for region %s ", region)
	}
	updatedTenants, err := tm.fetchTenantsForEventType(ctx, region, fromTimestamp, tm.supportedEventTypes.updatedTenantEvent)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching updated tenants for region %s ", region)
	}
	updatedTenants = excludeTenants(updatedTenants, fetchedTenants)
	fetchedTenants = append(fetchedTenants, updatedTenants...)
	return fetchedTenants, nil
}

func (tm *TenantsManager) TenantsToDelete(ctx context.Context, region, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	res, err := tm.fetchTenantsForEventType(ctx, region, fromTimestamp, tm.supportedEventTypes.deletedTenantEvent)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching deleted tenants for region %s ", region)
	}
	return res, nil
}

func (tm *TenantsManager) FetchTenant(ctx context.Context, externalTenantID string) (*model.BusinessTenantMappingInput, error) {
	additionalFields := map[string]string{
		tm.config.QueryConfig.EntityField: externalTenantID,
	}
	configProvider := eventsQueryConfigProviderWithAdditionalFields(tm.config, additionalFields)

	fetchedTenants, err := fetchCreatedTenantsWithRetries(tm.eventAPIClient, tm.config.RetryAttempts, tm.supportedEventTypes, configProvider)
	if err != nil {
		return nil, err
	}

	if len(fetchedTenants) >= 1 {
		log.C(ctx).Infof("Tenant found from central region with universal client")
		return &fetchedTenants[0], err
	}

	log.C(ctx).Infof("Tenant found from central region with universal client")

	tenantChan := make(chan *model.BusinessTenantMappingInput, len(tm.regionalClients))
	for _, regionDetails := range tm.regionalClients {
		go func(ctx context.Context, details RegionalClient, ch chan *model.BusinessTenantMappingInput) {
			createdRegionalTenants, err := fetchCreatedTenantsWithRetries(details.RegionalClient, tm.config.RetryAttempts, tm.supportedEventTypes, configProvider)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Failed to fetch created tenants from")
			}

			if len(createdRegionalTenants) == 1 {
				log.C(ctx).Infof("Tenant found in region %s", details.RegionName)
				ch <- &createdRegionalTenants[0]
			} else {
				log.C(ctx).Warnf("Tenant not found in region %s", details.RegionName)
				ch <- nil
			}
		}(ctx, regionDetails, tenantChan)
	}

	pendingRegionalInfo := len(tm.regionalClients)
	var tenant *model.BusinessTenantMappingInput
	for result := range tenantChan {
		if result != nil {
			tenant = result
			break
		}
		pendingRegionalInfo--
		if pendingRegionalInfo == 0 {
			// TODO return error when lazy store is reverted
			log.C(ctx).Error("tenant not found in all configured regions")
			return nil, nil
		}
	}

	return tenant, nil
}

func (tm *TenantsManager) CreateTenants(ctx context.Context, tenants []model.BusinessTenantMappingInput) error {
	tenantsToCreateGQL := tm.tenantConverter.MultipleInputToGraphQLInput(tenants)
	return runInChunks(ctx, tm.config.TenantOperationChunkSize, tenantsToCreateGQL, func(ctx context.Context, chunk []graphql.BusinessTenantMappingInput) error {
		return tm.gqlClient.WriteTenants(ctx, chunk)
	})
}

func (tm *TenantsManager) DeleteTenants(ctx context.Context, tenantsToDelete []model.BusinessTenantMappingInput) error {
	tenantsToDeleteGQL := tm.tenantConverter.MultipleInputToGraphQLInput(tenantsToDelete)
	return runInChunks(ctx, tm.config.TenantOperationChunkSize, tenantsToDeleteGQL, func(ctx context.Context, chunk []graphql.BusinessTenantMappingInput) error {
		return tm.gqlClient.DeleteTenants(ctx, chunk)
	})
}

func (tm *TenantsManager) fetchTenantsForEventType(ctx context.Context, region, fromTimestamp string, eventsType EventsType) ([]model.BusinessTenantMappingInput, error) {
	configProvider := eventsQueryConfigProviderWithRegion(tm.config, fromTimestamp, region)
	fetchedTenants, err := fetchTenantsWithRetries(tm.eventAPIClient, tm.config.RetryAttempts, eventsType, configProvider)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching tenants with universal client")
	}

	regionDetails, ok := tm.regionalClients[region]
	if !ok {
		log.C(ctx).Infof("Region %s does not have local events client enabled", region)
		return fetchedTenants, nil
	}
	configProvider = eventsQueryConfigProvider(tm.config, fromTimestamp)
	createdRegionalTenants, err := fetchTenantsWithRetries(regionDetails.RegionalClient, tm.config.RetryAttempts, eventsType, configProvider)
	if err != nil {
		return nil, err
	}
	createdRegionalTenants = excludeTenants(createdRegionalTenants, fetchedTenants)
	fetchedTenants = append(fetchedTenants, createdRegionalTenants...)

	return fetchedTenants, nil
}

func supportedEventTypes(tenantType tenantpkg.Type) (supportedEvents, error) {
	switch tenantType {
	case tenantpkg.Account:
		return supportedEvents{
			createdTenantEvent: CreatedAccountType,
			updatedTenantEvent: UpdatedAccountType,
			deletedTenantEvent: DeletedAccountType,
		}, nil
	case tenantpkg.Subaccount:
		return supportedEvents{
			createdTenantEvent: CreatedSubaccountType,
			updatedTenantEvent: UpdatedSubaccountType,
			deletedTenantEvent: DeletedSubaccountType,
		}, nil
	}
	return supportedEvents{}, errors.Errorf("Tenant events for type %s are not supported", tenantType)
}

func eventsQueryConfigProvider(config EventsConfig, fromTimestamp string) func() (QueryParams, PageConfig) {
	additionalFields := map[string]string{
		config.TimestampField: fromTimestamp,
	}
	return eventsQueryConfigProviderWithAdditionalFields(config, additionalFields)
}

func eventsQueryConfigProviderWithRegion(config EventsConfig, fromTimestamp, region string) func() (QueryParams, PageConfig) {
	additionalFields := map[string]string{
		config.QueryConfig.TimestampField: fromTimestamp,
	}
	if len(region) > 0 {
		additionalFields[config.QueryConfig.RegionField] = region
	}
	return eventsQueryConfigProviderWithAdditionalFields(config, additionalFields)
}

func eventsQueryConfigProviderWithAdditionalFields(config EventsConfig, additionalFields map[string]string) func() (QueryParams, PageConfig) {
	return func() (QueryParams, PageConfig) {
		qp := QueryParams{
			config.QueryConfig.PageNumField:  config.QueryConfig.PageStartValue,
			config.QueryConfig.PageSizeField: config.QueryConfig.PageSizeValue,
		}
		for field, value := range additionalFields {
			qp[field] = value
		}

		pc := PageConfig{
			TotalPagesField:   config.PagingConfig.TotalPagesField,
			TotalResultsField: config.PagingConfig.TotalResultsField,
			PageNumField:      config.QueryConfig.PageNumField,
		}
		return qp, pc
	}
}

func fetchCreatedTenantsWithRetries(eventAPIClient EventAPIClient, retryNumber uint, supportedEvents supportedEvents, configProvider func() (QueryParams, PageConfig)) ([]model.BusinessTenantMappingInput, error) {
	var fetchedTenants []model.BusinessTenantMappingInput

	createdTenants, err := fetchTenantsWithRetries(eventAPIClient, retryNumber, supportedEvents.createdTenantEvent, configProvider)
	if err != nil {
		return nil, fmt.Errorf("while fetching created tenants: %v", err)
	}
	fetchedTenants = append(fetchedTenants, createdTenants...)

	updatedTenants, err := fetchTenantsWithRetries(eventAPIClient, retryNumber, supportedEvents.updatedTenantEvent, configProvider)
	if err != nil {
		return nil, fmt.Errorf("while fetching updated tenants: %v", err)
	}

	updatedTenants = excludeTenants(updatedTenants, createdTenants)
	fetchedTenants = append(fetchedTenants, updatedTenants...)
	return fetchedTenants, nil
}

func fetchTenantsWithRetries(eventAPIClient EventAPIClient, retryNumber uint, eventsType EventsType, configProvider func() (QueryParams, PageConfig)) ([]model.BusinessTenantMappingInput, error) {
	var tenants []model.BusinessTenantMappingInput
	err := fetchWithRetries(retryNumber, func() error {
		fetchedTenants, err := fetchTenants(eventAPIClient, eventsType, configProvider)
		if err != nil {
			return fmt.Errorf("while fetching tenants: %v", err)
		}
		tenants = fetchedTenants
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func fetchTenants(eventAPIClient EventAPIClient, eventsType EventsType, configProvider func() (QueryParams, PageConfig)) ([]model.BusinessTenantMappingInput, error) {
	tenants := make([]model.BusinessTenantMappingInput, 0)
	err := walkThroughPages(eventAPIClient, eventsType, configProvider, func(page *EventsPage) error {
		mappings := page.getTenantMappings(eventsType)
		tenants = append(tenants, mappings...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("while walking through pages: %v", err)
	}

	return tenants, nil
}

func fetchWithRetries(retryAttempts uint, applyFunc func() error) error {
	err := retry.Do(applyFunc, retry.Attempts(retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))

	if err != nil {
		return err
	}
	return nil
}

func walkThroughPages(eventAPIClient EventAPIClient, eventsType EventsType, configProvider func() (QueryParams, PageConfig), applyFunc func(*EventsPage) error) error {
	params, pageConfig := configProvider()
	firstPage, err := eventAPIClient.FetchTenantEventsPage(eventsType, params)
	if err != nil {
		return errors.Wrap(err, "while fetching tenant events page")
	}
	if firstPage == nil {
		return nil
	}

	err = applyFunc(firstPage)
	if err != nil {
		return fmt.Errorf("while applyfunc on event page: %v", err)
	}

	initialCount := gjson.GetBytes(firstPage.Payload, pageConfig.TotalResultsField).Int()
	totalPages := gjson.GetBytes(firstPage.Payload, pageConfig.TotalPagesField).Int()

	pageStart, err := strconv.ParseInt(params[pageConfig.PageNumField], 10, 64)
	if err != nil {
		return err
	}

	for i := pageStart + 1; i <= totalPages; i++ {
		params[pageConfig.PageNumField] = strconv.FormatInt(i, 10)
		res, err := eventAPIClient.FetchTenantEventsPage(eventsType, params)
		if err != nil {
			return errors.Wrap(err, "while fetching tenant events page")
		}
		if res == nil {
			return apperrors.NewInternalError("next page was expected but response was empty")
		}
		if initialCount != gjson.GetBytes(res.Payload, pageConfig.TotalResultsField).Int() {
			return apperrors.NewInternalError("total results number changed during fetching consecutive events pages")
		}

		if err = applyFunc(res); err != nil {
			return err
		}
	}

	return nil
}

func runInChunks(ctx context.Context, maxChunkSize int, tenants []graphql.BusinessTenantMappingInput, storeTenantsFunc func(ctx context.Context, chunk []graphql.BusinessTenantMappingInput) error) error {
	for len(tenants) > 0 {
		chunkSize := int(math.Min(float64(len(tenants)), float64(maxChunkSize)))
		tenantsChunk := tenants[:chunkSize]
		if err := storeTenantsFunc(ctx, tenantsChunk); err != nil {
			return err
		}
		tenants = tenants[chunkSize:]
	}

	return nil
}
