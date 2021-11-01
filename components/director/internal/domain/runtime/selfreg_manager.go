package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// SelfRegConfig is configuration for the runtime self-registration flow
type SelfRegConfig struct {
	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`
	SelfRegisterLabelKey            string `envconfig:"APP_SELF_REGISTER_LABEL_KEY,optional"`
	SelfRegisterLabelValuePrefix    string `envconfig:"APP_SELF_REGISTER_LABEL_VALUE_PREFIX,optional"`
	SelfRegisterResponseKey         string `envconfig:"APP_SELF_REGISTER_RESPONSE_KEY,optional"`
	SelfRegisterPath                string `envconfig:"APP_SELF_REGISTER_PATH,optional"`
	SelfRegisterNameQueryParam      string `envconfig:"APP_SELF_REGISTER_NAME_QUERY_PARAM,optional"`
	SelfRegisterTenantQueryParam    string `envconfig:"APP_SELF_REGISTER_TENANT_QUERY_PARAM,optional"`
	SelfRegisterRequestBodyPattern  string `envconfig:"APP_SELF_REGISTER_REQUEST_BODY_PATTERN,optional"`

	ClientID     string `envconfig:"APP_SELF_REGISTER_CLIENT_ID,optional"`
	ClientSecret string `envconfig:"APP_SELF_REGISTER_CLIENT_SECRET,optional"`
	URL          string `envconfig:"APP_SELF_REGISTER_URL,optional"`

	ClientTimeout time.Duration `envconfig:"default=30s"`
}

// ExternalSvcCaller is used to call external services with given authentication
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore
type ExternalSvcCaller interface {
	Call(*http.Request) (*http.Response, error)
}

type selfRegisterManager struct {
	cfg    SelfRegConfig
	caller ExternalSvcCaller
}

// NewSelfRegisterManager creates a new SelfRegisterManager which is responsible for doing preparation/clean-up during
// self-registration of runtimes configured with values from cfg.
func NewSelfRegisterManager(cfg SelfRegConfig, caller ExternalSvcCaller) *selfRegisterManager {
	return &selfRegisterManager{cfg: cfg, caller: caller}
}

// PrepareRuntimeForSelfRegistration executes the prerequisite calls for self-registration in case the runtime
// is being self-registered
func (s *selfRegisterManager) PrepareRuntimeForSelfRegistration(ctx context.Context, in *graphql.RuntimeInput) error {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading consumer")
	}
	if distinguishLabel, exists := in.Labels[s.cfg.SelfRegisterDistinguishLabelKey]; exists && consumerInfo.Flow.IsCertFlow() { // this means that the runtime is being self-registered
		distinguishLabelVal := str.CastOrEmpty(distinguishLabel)

		request, err := s.createSelfRegPrepRequest(distinguishLabelVal, consumerInfo.ConsumerID)
		if err != nil {
			return err
		}

		response, err := s.caller.Call(request)
		if err != nil {
			return errors.Wrapf(err, "while executing preparation of self registered runtime")
		}
		defer httputils.Close(ctx, response.Body)

		respBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Wrapf(err, "while reading response body")
		}

		if response.StatusCode != http.StatusCreated {
			return errors.New(fmt.Sprintf("received unexpected status %d while preparing self-registered runtime: %s", response.StatusCode, string(respBytes)))
		}

		selfRegLabelVal := gjson.GetBytes(respBytes, s.cfg.SelfRegisterResponseKey)
		in.Labels[s.cfg.SelfRegisterLabelKey] = selfRegLabelVal.Str

		log.C(ctx).Infof("Successfully executed prep for self-registered runtime with distinguishing label value %s", distinguishLabelVal)
	}
	return nil
}

// CleanupSelfRegisteredRuntime executes cleanup calls for self-registered runtimes
func (s *selfRegisterManager) CleanupSelfRegisteredRuntime(ctx context.Context, selfRegisterDistinguishLabelValue string) error {
	if selfRegisterDistinguishLabelValue == "" {
		return nil
	}
	request, err := s.createSelfRegDelRequest(selfRegisterDistinguishLabelValue)
	if err != nil {
		return err
	}

	resp, err := s.caller.Call(request)
	if err != nil {
		return errors.Wrapf(err, "while executing cleanup of self-registered runtime with label value %s", selfRegisterDistinguishLabelValue)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("received unexpected status code %d while cleaning up self-registered runtime with label value %s", resp.StatusCode, selfRegisterDistinguishLabelValue))
	}

	log.C(ctx).Infof("Successfully executed clean-up self-registered runtime with distinguishing label value %s", selfRegisterDistinguishLabelValue)
	return nil
}

// GetSelfRegDistinguishingLabelKey returns the label key to be used in order to determine whether a runtime
// is being self-registered.
func (s *selfRegisterManager) GetSelfRegDistinguishingLabelKey() string {
	return s.cfg.SelfRegisterDistinguishLabelKey
}

func (s *selfRegisterManager) createSelfRegPrepRequest(distinguishingVal, tenant string) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + distinguishingVal
	url, err := urlpkg.Parse(s.cfg.URL + s.cfg.SelfRegisterPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for preparation of self-registered runtime")
	}

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

func (s *selfRegisterManager) createSelfRegDelRequest(distinguishingVal string) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + distinguishingVal
	url, err := urlpkg.Parse(s.cfg.URL + s.cfg.SelfRegisterPath + selfRegLabelVal)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for cleanup of self-registered runtime")
	}

	request, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing request for cleanup of self-registered runtime")
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}
