package tenantfetchersvc

import (
	"context"
	"fmt"

	labelutils "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// TenantProvisioner is used to create all related to the incoming request tenants, and build their hierarchy;
//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore
type TenantProvisioner interface {
	ProvisionTenants(context.Context, *TenantSubscriptionRequest, string) error
}

// RuntimeService is used to interact with runtimes
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	SetLabel(context.Context, *model.LabelInput) error
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
	ListByFiltersGlobal(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

type subscriptionFunc func(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error

type sliceMutationFunc func([]string, string) []string

type subscriber struct {
	provisioner TenantProvisioner
	RuntimeService
	SubscriptionProviderLabelKey  string
	ConsumerSubaccountIDsLabelKey string
}

// NewSubscriber creates new subscriber
func NewSubscriber(provisioner TenantProvisioner, service RuntimeService, subscriptionProviderLabelKey, consumerSubaccountIDsLabelKey string) *subscriber {
	return &subscriber{
		provisioner:                   provisioner,
		RuntimeService:                service,
		SubscriptionProviderLabelKey:  subscriptionProviderLabelKey,
		ConsumerSubaccountIDsLabelKey: consumerSubaccountIDsLabelKey,
	}
}

// Subscribe subscribes tenant to runtime. If the tenant does not exist it will be created
func (s *subscriber) Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error {
	if err := s.provisioner.ProvisionTenants(ctx, tenantSubscriptionRequest, region); err != nil {
		return err
	}

	return s.applyRuntimesSubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionConsumerID, tenantSubscriptionRequest.SubaccountTenantID, region, addElement)
}

// Unsubscribe unsubscribes tenant from runtime.
func (s *subscriber) Unsubscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error {
	return s.applyRuntimesSubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionConsumerID, tenantSubscriptionRequest.SubaccountTenantID, region, removeElement)
}

func (s *subscriber) applyRuntimesSubscriptionChange(ctx context.Context, subscriptionConsumerID, subaccountTenantID, region string, mutateLabelsFunc sliceMutationFunc) error {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.SubscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", subscriptionConsumerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	runtimes, err := s.ListByFiltersGlobal(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil
		}

		return errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", tenant.RegionLabelKey, region, s.SubscriptionProviderLabelKey, subscriptionConsumerID))
	}

	for _, runtime := range runtimes {
		ctx = tenant.SaveToContext(ctx, runtime.Tenant, "")

		labelOldValue := make([]string, 0)
		label, err := s.GetLabel(ctx, runtime.ID, s.ConsumerSubaccountIDsLabelKey)
		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", runtime.ID, s.ConsumerSubaccountIDsLabelKey))
			}
			// if the error is not found, do nothing and continue
		} else {
			if labelOldValue, err = labelutils.ValueToStringsSlice(label.Value); err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
		}

		var labelNewValue = mutateLabelsFunc(labelOldValue, subaccountTenantID)

		if err := s.SetLabel(ctx, &model.LabelInput{
			Key:        s.ConsumerSubaccountIDsLabelKey,
			Value:      labelNewValue,
			ObjectType: model.RuntimeLabelableObject,
			ObjectID:   runtime.ID,
		}); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", runtime.ID))
		}
	}

	return nil
}

func addElement(slice []string, elem string) []string {
	return append(slice, elem)
}

func removeElement(slice []string, elem string) []string {
	result := make([]string, 0)
	for _, e := range slice {
		if e != elem {
			result = append(result, e)
		}
	}
	return result
}
