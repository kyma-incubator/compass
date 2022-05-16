package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"path"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/config"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// ExternalSvcCaller is used to call external services with given authentication
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCaller interface {
	Call(*http.Request) (*http.Response, error)
}

// ExternalSvcCallerProvider provides ExternalSvcCaller based on the provided SelfRegConfig and region
//go:generate mockery --name=ExternalSvcCallerProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCallerProvider interface {
	GetCaller(config.SelfRegConfig, string) (ExternalSvcCaller, error)
}

const regionLabel = "region"

type selfRegisterManager struct {
	cfg            config.SelfRegConfig
	callerProvider ExternalSvcCallerProvider
}

// NewSelfRegisterManager creates a new SelfRegisterManager which is responsible for doing preparation/clean-up during
// self-registration of runtimes configured with values from cfg.
func NewSelfRegisterManager(cfg config.SelfRegConfig, provider ExternalSvcCallerProvider) (*selfRegisterManager, error) {
	if err := cfg.MapInstanceConfigs(); err != nil {
		return nil, errors.Wrap(err, "while creating self register manager")
	}
	return &selfRegisterManager{cfg: cfg, callerProvider: provider}, nil
}

// PrepareRuntimeForSelfRegistration executes the prerequisite calls for self-registration in case the runtime
// is being self-registered
func (s *selfRegisterManager) PrepareRuntimeForSelfRegistration(ctx context.Context, in model.RuntimeInput, id string) (map[string]interface{}, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	labels := make(map[string]interface{})
	if err != nil {
		return labels, errors.Wrapf(err, "while loading consumer")
	}
	if distinguishLabel, exists := in.Labels[s.cfg.SelfRegisterDistinguishLabelKey]; exists && consumerInfo.Flow.IsCertFlow() { // this means that the runtime is being self-registered
		regionValue, exists := in.Labels[regionLabel]
		if !exists {
			return labels, errors.Errorf("missing %q label", regionLabel)
		}

		region, ok := regionValue.(string)
		if !ok {
			return labels, errors.Errorf("region value should be of type %q", "string")
		}

		instanceConfig, exists := s.cfg.RegionToInstanceConfig[region]
		if !exists {
			return labels, errors.Errorf("missing configuration for region: %s", region)
		}

		request, err := s.createSelfRegPrepRequest(id, consumerInfo.ConsumerID, instanceConfig.URL)
		if err != nil {
			return labels, err
		}

		caller, err := s.callerProvider.GetCaller(s.cfg, region)
		if err != nil {
			return labels, errors.Wrapf(err, "while getting caller")
		}

		response, err := caller.Call(request)
		if err != nil {
			return labels, errors.Wrapf(err, "while executing preparation of self registered runtime")
		}
		defer httputils.Close(ctx, response.Body)

		respBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return labels, errors.Wrapf(err, "while reading response body")
		}

		if response.StatusCode != http.StatusCreated {
			return labels, apperrors.NewCustomErrorWithCode(response.StatusCode, fmt.Sprintf("received unexpected status %d while preparing self-registered runtime: %s", response.StatusCode, string(respBytes)))
		}

		selfRegLabelVal := gjson.GetBytes(respBytes, s.cfg.SelfRegisterResponseKey)
		labels[s.cfg.SelfRegisterLabelKey] = selfRegLabelVal.Str

		log.C(ctx).Infof("Successfully executed prep for self-registered runtime with distinguishing label value %s", str.CastOrEmpty(distinguishLabel))
	}
	return labels, nil
}

// CleanupSelfRegisteredRuntime executes cleanup calls for self-registered runtimes
func (s *selfRegisterManager) CleanupSelfRegisteredRuntime(ctx context.Context, runtimeID, region string) error {
	if runtimeID == "" {
		return nil
	}

	instanceConfig, exists := s.cfg.RegionToInstanceConfig[region]
	if !exists {
		return errors.Errorf("missing configuration for region: %s", region)
	}

	request, err := s.createSelfRegDelRequest(runtimeID, instanceConfig.URL)
	if err != nil {
		return err
	}

	caller, err := s.callerProvider.GetCaller(s.cfg, region)
	if err != nil {
		return errors.Wrapf(err, "while getting caller")
	}
	resp, err := caller.Call(request)
	if err != nil {
		return errors.Wrapf(err, "while executing cleanup of self-registered runtime with id %q", runtimeID)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received unexpected status code %d while cleaning up self-registered runtime with id %q", resp.StatusCode, runtimeID)
	}

	log.C(ctx).Infof("Successfully executed clean-up self-registered runtime with id %q", runtimeID)
	return nil
}

// GetSelfRegDistinguishingLabelKey returns the label key to be used in order to determine whether a runtime
// is being self-registered.
func (s *selfRegisterManager) GetSelfRegDistinguishingLabelKey() string {
	return s.cfg.SelfRegisterDistinguishLabelKey
}

func (s *selfRegisterManager) createSelfRegPrepRequest(runtimeID, tenant, targetURL string) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + runtimeID
	url, err := urlpkg.Parse(targetURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for preparation of self-registered runtime")
	}
	url.Path = path.Join(url.Path, s.cfg.SelfRegisterPath)
	q := url.Query()
	q.Add(s.cfg.SelfRegisterNameQueryParam, selfRegLabelVal)
	q.Add(s.cfg.SelfRegisterTenantQueryParam, tenant)
	url.RawQuery = q.Encode()

	request, err := http.NewRequest(http.MethodPost, url.String(), strings.NewReader(fmt.Sprintf(s.cfg.SelfRegisterRequestBodyPattern, selfRegLabelVal)))
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing request for self-registered runtime")
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func (s *selfRegisterManager) createSelfRegDelRequest(runtimeID, targetURL string) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + runtimeID
	url, err := urlpkg.Parse(targetURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for cleanup of self-registered runtime")
	}
	url.Path = path.Join(url.Path, s.cfg.SelfRegisterPath)
	url.Path = path.Join(url.Path, selfRegLabelVal)

	request, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing request for cleanup of self-registered runtime")
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}
