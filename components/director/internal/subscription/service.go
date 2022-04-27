package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	labelPkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// Config is configuration for the tenant-runtime subscription flow
type Config struct {
	ProviderLabelKey              string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountIDsLabelKey string `envconfig:"APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY,default=consumer_subaccount_ids"`
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	ListAll(context.Context, string, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// TenantService provides functionality for retrieving, and creating tenants.
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --unroll-variadic=False
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
}

// LabelService is responsible updating already existing labels, and their label definitions.
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore
type LabelService interface {
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore
type uidService interface {
	Generate() string
}

const (
	retryAttempts          = 2
	retryDelayMilliseconds = 100
)

type service struct {
	runtimeSvc                    RuntimeService
	tenantSvc                     TenantService
	labelSvc                      LabelService
	uidSvc                        uidService
	subscriptionProviderLabelKey  string
	consumerSubaccountIDsLabelKey string
}

// NewService missing godoc
func NewService(runtimeSvc RuntimeService, tenantSvc TenantService, labelSvc LabelService, uidService uidService,
	subscriptionProviderLabelKey string, consumerSubaccountIDsLabelKey string) *service {
	return &service{
		runtimeSvc:                    runtimeSvc,
		tenantSvc:                     tenantSvc,
		labelSvc:                      labelSvc,
		uidSvc:                        uidService,
		subscriptionProviderLabelKey:  subscriptionProviderLabelKey,
		consumerSubaccountIDsLabelKey: consumerSubaccountIDsLabelKey,
	}
}

func (s *service) SubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error) {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", providerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	runtimes, err := s.runtimeSvc.ListAll(ctx, providerSubaccountID, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}

	for _, provider := range runtimes {
		tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, provider.ID)
		if err != nil {
			return false, err
		}

		label, err := s.labelSvc.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.consumerSubaccountIDsLabelKey,
			ObjectID:   provider.ID,
			ObjectType: model.RuntimeLabelableObject,
		})

		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", provider.ID, s.consumerSubaccountIDsLabelKey))
			}
			if err := s.createLabel(ctx, tnt, provider, subaccountTenantID); err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to create label with key: %s", s.consumerSubaccountIDsLabelKey))
			}
		} else {
			labelOldValue, err := labelPkg.ValueToStringsSlice(label.Value)
			if err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
			labelNewValue := append(labelOldValue, subaccountTenantID)

			if err := s.updateLabelWithRetry(ctx, tnt, provider, label, labelNewValue); err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", provider.ID))
			}
		}
	}
	return true, nil
}

func (s *service) UnsubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error) {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", providerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	runtimes, err := s.runtimeSvc.ListAll(ctx, providerSubaccountID, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}

	for _, runtime := range runtimes {
		tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, runtime.ID)
		if err != nil {
			return false, err
		}

		label, err := s.labelSvc.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.consumerSubaccountIDsLabelKey,
			ObjectID:   runtime.ID,
			ObjectType: model.RuntimeLabelableObject,
		})

		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", runtime.ID, s.consumerSubaccountIDsLabelKey))
			}
			return true, nil
		} else {
			labelOldValue, err := labelPkg.ValueToStringsSlice(label.Value)
			if err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
			labelNewValue := removeElement(labelOldValue, subaccountTenantID)

			if err := s.updateLabelWithRetry(ctx, tnt, runtime, label, labelNewValue); err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", runtime.ID))
			}
		}
	}

	return true, nil
}

func (s *service) createLabel(ctx context.Context, tenant string, runtime *model.Runtime, subaccountTenantID string) error {
	return s.labelSvc.CreateLabel(ctx, tenant, s.uidSvc.Generate(), &model.LabelInput{
		Key:        s.consumerSubaccountIDsLabelKey,
		Value:      []string{subaccountTenantID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtime.ID,
	})
}

func (s *service) updateLabelWithRetry(ctx context.Context, tenant string, runtime *model.Runtime, label *model.Label, labelNewValue []string) error {
	return retry.Do(func() error {
		err := s.labelSvc.UpdateLabel(ctx, tenant, label.ID, &model.LabelInput{
			Key:        s.consumerSubaccountIDsLabelKey,
			Value:      labelNewValue,
			ObjectType: model.RuntimeLabelableObject,
			ObjectID:   runtime.ID,
			Version:    label.Version,
		})
		if err != nil {
			return errors.Wrap(err, "while updating label")
		}
		return nil
	}, retry.Attempts(retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))
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
