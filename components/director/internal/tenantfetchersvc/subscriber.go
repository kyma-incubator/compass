package tenantfetchersvc

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	ListByFiltersGlobal(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// LabelService is responsible updating already existing labels, and their label definitions.
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore
type LabelService interface {
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
}

// UIDService generates a unique ID
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type subscriptionFunc func(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error

type sliceMutationFunc func([]string, string) []string

type subscriber struct {
	provisioner                   TenantProvisioner
	runtimeSvc                    RuntimeService
	labelSvc                      LabelService
	uidSvc                        UIDService
	tenantSvc                     TenantService
	SubscriptionProviderLabelKey  string
	ConsumerSubaccountIDsLabelKey string
}

// NewSubscriber creates new subscriber
func NewSubscriber(provisioner TenantProvisioner, service RuntimeService, labelService LabelService, uuidSvc UIDService, tenantSvc TenantService, subscriptionProviderLabelKey, consumerSubaccountIDsLabelKey string) *subscriber {
	return &subscriber{
		provisioner:                   provisioner,
		runtimeSvc:                    service,
		labelSvc:                      labelService,
		uidSvc:                        uuidSvc,
		tenantSvc:                     tenantSvc,
		SubscriptionProviderLabelKey:  subscriptionProviderLabelKey,
		ConsumerSubaccountIDsLabelKey: consumerSubaccountIDsLabelKey,
	}
}

// Subscribe subscribes tenant to runtime. If the tenant does not exist it will be created
func (s *subscriber) Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error {
	if err := s.provisioner.ProvisionTenants(ctx, tenantSubscriptionRequest, region); err != nil {
		return err
	}

	return s.applyRuntimesSubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.SubaccountTenantID, region, addElement)
}

// Unsubscribe unsubscribes tenant from runtime.
func (s *subscriber) Unsubscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error {
	return s.applyRuntimesSubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.SubaccountTenantID, region, removeElement)
}

func (s *subscriber) applyRuntimesSubscriptionChange(ctx context.Context, subscriptionProviderID, subaccountTenantID, region string, mutateLabelsFunc sliceMutationFunc) error {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.SubscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	runtimes, err := s.runtimeSvc.ListByFiltersGlobal(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil
		}

		return errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", tenant.RegionLabelKey, region, s.SubscriptionProviderLabelKey, subscriptionProviderID))
	}

	for _, runtime := range runtimes {
		tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, runtime.ID)
		if err != nil {
			return err
		}

		label, err := s.labelSvc.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.ConsumerSubaccountIDsLabelKey,
			ObjectID:   runtime.ID,
			ObjectType: model.RuntimeLabelableObject,
		})

		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", runtime.ID, s.ConsumerSubaccountIDsLabelKey))
			}
			// if the error is not found, create a label
			if err := s.createLabel(ctx, tnt, runtime, subaccountTenantID); err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to create label with key: %s", s.ConsumerSubaccountIDsLabelKey))
			}
		} else {
			labelOldValue, err := labelutils.ValueToStringsSlice(label.Value)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
			labelNewValue := mutateLabelsFunc(labelOldValue, subaccountTenantID)

			if err := s.updateLabel(ctx, tnt, runtime, label, labelNewValue); err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", runtime.ID))
			}
		}
	}

	return nil
}

func (s *subscriber) createLabel(ctx context.Context, tenant string, runtime *model.Runtime, subaccountTenantID string) error {
	return s.labelSvc.CreateLabel(ctx, tenant, s.uidSvc.Generate(), &model.LabelInput{
		Key:        s.ConsumerSubaccountIDsLabelKey,
		Value:      []string{subaccountTenantID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtime.ID,
	})
}

func (s *subscriber) updateLabel(ctx context.Context, tenant string, runtime *model.Runtime, label *model.Label, labelNewValue []string) error {
	return s.labelSvc.UpdateLabel(ctx, tenant, label.ID, &model.LabelInput{
		Key:        s.ConsumerSubaccountIDsLabelKey,
		Value:      labelNewValue,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtime.ID,
		Version:    label.Version,
	})
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
