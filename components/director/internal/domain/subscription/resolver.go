package subscription

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/apiclient"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const subscriptionIDKey = "subscriptionGUID"

// DependentServiceInstanceInfo represents the dependent service instance info object in a subscription payload.
type DependentServiceInstanceInfo struct {
	AppID                string `json:"appId"`
	AppName              string `json:"appName"`
	ProviderSubaccountID string `json:"providerSubaccountId"`
}

// DependentServiceInstancesInfo represents collection of all dependent service instance info objects in a subscription payload.
type DependentServiceInstancesInfo struct {
	Instances []DependentServiceInstanceInfo `json:"dependentServiceInstancesInfo"`
}

// SubscriptionService responsible for service-layer Subscription operations
//
//go:generate mockery --name=SubscriptionService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SubscriptionService interface {
	SubscribeTenantToRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID string) (bool, error)
	UnsubscribeTenantFromRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID string) (bool, error)
	SubscribeTenantToApplication(ctx context.Context, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID string, subscriptionPayload string) (bool, string, string, error)
	UnsubscribeTenantFromApplication(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID string) (bool, error)
	DetermineSubscriptionFlow(ctx context.Context, providerID, region string) (resource.Type, error)
}

// Resolver is an object responsible for resolver-layer Subscription operations.
type Resolver struct {
	transact        persistence.Transactioner
	subscriptionSvc SubscriptionService
	ordClient       *apiclient.ORDClient
}

// NewResolver returns a new object responsible for resolver-layer Subscription operations.
func NewResolver(transact persistence.Transactioner, subscriptionSvc SubscriptionService, ordAggregatorClientConfig apiclient.OrdAggregatorClientConfig) *Resolver {
	return &Resolver{
		transact:        transact,
		subscriptionSvc: subscriptionSvc,
		ordClient:       apiclient.NewORDClient(ordAggregatorClientConfig),
	}
}

// SubscribeTenant subscribes tenant to runtime or application template labeled with `providerID` and `region`
func (r *Resolver) SubscribeTenant(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName string, subscriptionPayload string) (bool, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return false, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var dependentSvcInstancesInfo DependentServiceInstancesInfo
	if err = json.Unmarshal([]byte(subscriptionPayload), &dependentSvcInstancesInfo); err != nil {
		return false, errors.Wrapf(err, "while unmarshaling dependent service instances info")
	}
	var success bool

	appID := ""
	appTemplateID := ""
	var flowType resource.Type
	for _, instance := range dependentSvcInstancesInfo.Instances {
		log.C(ctx).Infof("Subscription flow for subscribe will be entered. Changing provider ID from %q to %q, provider subaccount id from %q to %q and subscription app name from %q to %q", providerID, instance.AppID, providerSubaccountID, instance.ProviderSubaccountID, subscriptionAppName, instance.AppName)

		providerID = instance.AppID
		if providerID == "" {
			return false, errors.New("Provider ID should not be empty")
		}

		providerSubaccountID = instance.ProviderSubaccountID
		if providerSubaccountID == "" {
			return false, errors.New("Provider subaccount ID should not be empty")
		}

		subscriptionAppName = instance.AppName
		if subscriptionAppName == "" {
			return false, errors.New("Subscription app name should not be empty")
		}

		subscriptionID := gjson.GetBytes([]byte(subscriptionPayload), subscriptionIDKey).String()
		if subscriptionID == "" {
			return false, errors.New("Subscription ID should not be empty")
		}

		flowType, err = r.subscriptionSvc.DetermineSubscriptionFlow(ctx, providerID, region)
		if err != nil {
			return false, errors.Wrap(err, "while determining subscription flow")
		}

		switch flowType {
		case resource.ApplicationTemplate:
			log.C(ctx).Infof("Entering application subscription flow")
			success, appID, appTemplateID, err = r.subscriptionSvc.SubscribeTenantToApplication(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID, subscriptionPayload)
			if err != nil {
				return false, err
			}
		case resource.Runtime:
			log.C(ctx).Infof("Entering runtime subscription flow")
			success, err = r.subscriptionSvc.SubscribeTenantToRuntime(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID)
			if err != nil {
				return false, err
			}
		default:
			log.C(ctx).Infof("Nothing to subscribe to provider (%q) in region (%q)", providerID, region)
		}
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	if flowType == resource.ApplicationTemplate && appID != "" && appTemplateID != "" {
		if err := r.ordClient.Aggregate(ctx, appID, appTemplateID); err != nil {
			log.C(ctx).WithError(err).Errorf("Error while calling aggregate API with AppID %q and AppTemplateID %q", appID, appTemplateID)
		}
		if err := r.ordClient.Aggregate(ctx, appID, ""); err != nil {
			log.C(ctx).WithError(err).Errorf("Error while calling aggregate API with AppID %q", appID)
		}
	}

	return success, nil
}

// UnsubscribeTenant unsubscribes tenant from runtime or application template labeled with `providerID` and `region`
func (r *Resolver) UnsubscribeTenant(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionPayload string) (bool, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return false, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var dependentSvcInstancesInfo DependentServiceInstancesInfo
	if err = json.Unmarshal([]byte(subscriptionPayload), &dependentSvcInstancesInfo); err != nil {
		return false, errors.Wrapf(err, "while unmarshaling dependent service instances info")
	}

	var success bool

	for _, instance := range dependentSvcInstancesInfo.Instances {
		log.C(ctx).Infof("Subscription flow for unsubscribe will be entered. Changing provider ID from %q to %q and provider subaccount id from %q to %q", providerID, instance.AppID, providerSubaccountID, instance.ProviderSubaccountID)
		providerID = instance.AppID
		providerSubaccountID = instance.ProviderSubaccountID
		subscriptionID := gjson.GetBytes([]byte(subscriptionPayload), subscriptionIDKey).String()

		flowType, err := r.subscriptionSvc.DetermineSubscriptionFlow(ctx, providerID, region)
		if err != nil {
			return false, errors.Wrap(err, "while determining subscription flow")
		}

		switch flowType {
		case resource.ApplicationTemplate:
			log.C(ctx).Infof("Entering application subscription flow")
			success, err = r.subscriptionSvc.UnsubscribeTenantFromApplication(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)
			if err != nil {
				return false, err
			}
		case resource.Runtime:
			log.C(ctx).Infof("Entering runtime subscription flow")
			success, err = r.subscriptionSvc.UnsubscribeTenantFromRuntime(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)
			if err != nil {
				return false, err
			}
		default:
			log.C(ctx).Infof("Nothing to unsubscribe from provider(%q) with subaccount: %q in region (%q)", providerID, providerSubaccountID, region)
		}
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return success, nil
}
