package selfregmanager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"path"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/director/pkg/config"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// ExternalSvcCaller is used to call external services with given authentication
//
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCaller interface {
	Call(*http.Request) (*http.Response, error)
}

// ExternalSvcCallerProvider provides ExternalSvcCaller based on the provided SelfRegConfig and region
//
//go:generate mockery --name=ExternalSvcCallerProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCallerProvider interface {
	GetCaller(config.SelfRegConfig, string) (ExternalSvcCaller, error)
}

// RegionLabel label for the label repository indicating region
const RegionLabel = "region"

type selfRegisterManager struct {
	cfg            config.SelfRegConfig
	callerProvider ExternalSvcCallerProvider
}

// NewSelfRegisterManager creates a new SelfRegisterManager which is responsible for doing preparation/clean-up during
// self-registration of runtimes configured with values from cfg.
func NewSelfRegisterManager(cfg config.SelfRegConfig, provider ExternalSvcCallerProvider) (*selfRegisterManager, error) {
	if err := cfg.PrepareConfiguration(); err != nil {
		return nil, errors.Wrap(err, "while preparing self register manager configuration")
	}
	return &selfRegisterManager{cfg: cfg, callerProvider: provider}, nil
}

// IsSelfRegistrationFlow check if self registration flow is triggered
func (s *selfRegisterManager) IsSelfRegistrationFlow(ctx context.Context, labels map[string]interface{}) (bool, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading consumer")
	}

	if consumerInfo.Flow.IsCertFlow() {
		if _, exists := labels[s.cfg.SelfRegisterDistinguishLabelKey]; !exists {
			return false, errors.Errorf("missing %q label", s.cfg.SelfRegisterDistinguishLabelKey)
		}

		return true, nil
	}
	return false, nil
}

// PrepareForSelfRegistration executes the prerequisite calls for self-registration in case the runtime
// is being self-registered
func (s *selfRegisterManager) PrepareForSelfRegistration(ctx context.Context, resourceType resource.Type, labels map[string]interface{}, id string, validate func() error) (map[string]interface{}, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading consumer")
	}

	if consumerInfo.Flow.IsCertFlow() {
		distinguishLabel, exists := labels[s.cfg.SelfRegisterDistinguishLabelKey]
		if !exists {
			if resourceType == resource.Runtime {
				return labels, nil
			}
			return nil, errors.Errorf("missing %q label", s.cfg.SelfRegisterDistinguishLabelKey)
		}

		if err := validate(); err != nil {
			return nil, err
		}

		if labels[RegionLabel] != nil {
			return nil, errors.Errorf("providing %q label and value is forbidden", RegionLabel)
		}

		region := consumerInfo.Region
		if region == "" {
			return nil, errors.Errorf("missing %s value in consumer context", RegionLabel)
		}

		labels[RegionLabel] = region

		instanceConfig, exists := s.cfg.RegionToInstanceConfig[region]
		if !exists {
			return nil, errors.Errorf("missing configuration for region: %s", region)
		}

		request, err := s.createSelfRegPrepRequest(id, consumerInfo.ConsumerID, instanceConfig.URL)
		if err != nil {
			return nil, err
		}

		caller, err := s.callerProvider.GetCaller(s.cfg, region)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting caller")
		}

		response, err := caller.Call(request)
		if err != nil {
			return nil, errors.Wrapf(err, "while executing preparation of self registered resource")
		}
		defer httputils.Close(ctx, response.Body)

		respBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "while reading response body")
		}

		if response.StatusCode != http.StatusCreated {
			return nil, apperrors.NewCustomErrorWithCode(response.StatusCode, fmt.Sprintf("received unexpected status %d while preparing self-registered resource: %s", response.StatusCode, string(respBytes)))
		}

		selfRegLabelVal := gjson.GetBytes(respBytes, s.cfg.SelfRegisterResponseKey)
		labels[s.cfg.SelfRegisterLabelKey] = selfRegLabelVal.Str

		if resourceType == resource.Runtime {
			saasAppName, exists := s.cfg.RegionToSaaSAppName[region]
			if !exists {
				return nil, errors.Errorf("missing SaaS application name for region: %q", region)
			}

			if saasAppName == "" {
				return nil, errors.Errorf("SaaS application name for region: %q could not be empty", region)
			}

			labels[s.cfg.SaaSAppNameLabelKey] = saasAppName
		}

		if resourceType == resource.ApplicationTemplate {
			labels[scenarioassignment.SubaccountIDKey] = consumerInfo.ConsumerID
		}

		log.C(ctx).Infof("Successfully executed prep for self-registration with distinguishing label value %s", str.CastOrEmpty(distinguishLabel))
	}

	return labels, nil
}

// CleanupSelfRegistration executes cleanup calls for self-registered runtimes
func (s *selfRegisterManager) CleanupSelfRegistration(ctx context.Context, resourceID, region string) error {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading consumer")
	}

	if !consumerInfo.Flow.IsCertFlow() {
		log.C(ctx).Infof("Not certificate flow, skipping clone deletion for resource %q", resourceID)
		return nil
	}

	if resourceID == "" {
		return nil
	}

	instanceConfig, exists := s.cfg.RegionToInstanceConfig[region]
	if !exists {
		return errors.Errorf("missing configuration for region: %s", region)
	}

	request, err := s.createSelfRegDelRequest(resourceID, instanceConfig.URL)
	if err != nil {
		return err
	}

	caller, err := s.callerProvider.GetCaller(s.cfg, region)
	if err != nil {
		return errors.Wrapf(err, "while getting caller")
	}
	resp, err := caller.Call(request)
	if err != nil {
		return errors.Wrapf(err, "while executing cleanup of self-registered resource with id %q", resourceID)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received unexpected status code %d while cleaning up self-registered resource with id %q", resp.StatusCode, resourceID)
	}

	log.C(ctx).Infof("Successfully executed clean-up self-registered resource with id %q", resourceID)
	return nil
}

// GetSelfRegDistinguishingLabelKey returns the label key to be used in order to determine whether a resource
// is being self-registered.
func (s *selfRegisterManager) GetSelfRegDistinguishingLabelKey() string {
	return s.cfg.SelfRegisterDistinguishLabelKey
}

func (s *selfRegisterManager) createSelfRegPrepRequest(id, tenant, targetURL string) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + id
	url, err := urlpkg.Parse(targetURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for preparation of self-registered resource")
	}
	url.Path = path.Join(url.Path, s.cfg.SelfRegisterPath)
	q := url.Query()
	q.Add(s.cfg.SelfRegisterNameQueryParam, selfRegLabelVal)
	q.Add(s.cfg.SelfRegisterTenantQueryParam, tenant)
	url.RawQuery = q.Encode()

	request, err := http.NewRequest(http.MethodPost, url.String(), strings.NewReader(fmt.Sprintf(s.cfg.SelfRegisterRequestBodyPattern, selfRegLabelVal)))
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing request for self-registered resource")
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func (s *selfRegisterManager) createSelfRegDelRequest(resourceID, targetURL string) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + resourceID
	url, err := urlpkg.Parse(targetURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for cleanup of self-registered resource")
	}
	url.Path = path.Join(url.Path, s.cfg.SelfRegisterPath)
	url.Path = path.Join(url.Path, selfRegLabelVal)

	request, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing request for cleanup of self-registered resource")
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}
