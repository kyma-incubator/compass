package subscription

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// SubscriptionService responsible for service-layer Subscription operations
//go:generate mockery --name=SubscriptionService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SubscriptionService interface {
	SubscribeTenantToRuntime(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error)
	UnsubscribeTenantFromRuntime(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error)
	SubscribeTenantToApplication(ctx context.Context, subaccountTenantID, region, providerSubaccountID, subscribedSubaccountID, subscribedAppName string) (bool, error)
	UnsubscribeTenantFromApplication(ctx context.Context, subaccountTenantID, region, providerSubaccountID string) (bool, error)
	DetermineSubscriptionFlow(ctx context.Context, providerID, region string) (resource.Type, error)
}

// Resolver is an object responsible for resolver-layer Subscription operations.
type Resolver struct {
	transact        persistence.Transactioner
	subscriptionSvc SubscriptionService
}

// NewResolver returns a new object responsible for resolver-layer Subscription operations.
func NewResolver(transact persistence.Transactioner, subscriptionSvc SubscriptionService) *Resolver {
	return &Resolver{
		transact:        transact,
		subscriptionSvc: subscriptionSvc,
	}
}

// SubscribeTenant subscribes tenant to runtime labeled with `providerID` and `region`
func (r *Resolver) SubscribeTenant(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, region, subscriptionAppName string) (bool, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return false, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	flowType, err := r.subscriptionSvc.DetermineSubscriptionFlow(ctx, providerID, region)
	if err != nil {
		return false, errors.Wrap(err, "while determining subscription flow")
	}

	var success bool

	if flowType == resource.ApplicationTemplate {
		log.C(ctx).Infof("Entering Application flow")
		success, err = r.subscriptionSvc.SubscribeTenantToApplication(ctx, providerID, region, providerSubaccountID, subaccountTenantID, subscriptionAppName)
		if err != nil {
			return false, err
		}
	} else {
		log.C(ctx).Infof("Entering Runtime flow")
		success, err = r.subscriptionSvc.SubscribeTenantToRuntime(ctx, providerID, subaccountTenantID, providerSubaccountID, region)
		if err != nil {
			return false, err
		}
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return success, nil
}

// UnsubscribeTenant unsubscribes tenant to runtime labeled with `providerID` and `region`
func (r *Resolver) UnsubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return false, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	flowType, err := r.subscriptionSvc.DetermineSubscriptionFlow(ctx, providerID, region)
	if err != nil {
		return false, errors.Wrap(err, "while determining subscription flow")
	}

	var success bool

	if flowType == resource.ApplicationTemplate {
		log.C(ctx).Infof("Entering Application flow")
		success, err = r.subscriptionSvc.UnsubscribeTenantFromApplication(ctx, providerID, region, providerSubaccountID)
		if err != nil {
			return false, err
		}
	} else {
		log.C(ctx).Infof("Entering Runtime flow")
		success, err = r.subscriptionSvc.UnsubscribeTenantFromRuntime(ctx, providerID, subaccountTenantID, providerSubaccountID, region)
		if err != nil {
			return false, err
		}
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return success, nil
}
