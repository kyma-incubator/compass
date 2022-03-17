package runtime

import "C"
import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/securehttp"
	authpkg "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"io"
	"net/http"
	urlpkg "net/url"
	"path"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// ExternalSvcCallerProvider is used to call external services with given authentication
//go:generate mockery --name=ExternalSvcCallerProvider --output=automock --outpkg=automock --case=underscore
type ExternalSvcCallerProvider interface {
	GetCaller(config SelfRegConfig, region string) (*securehttp.Caller, error)
}


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

	OAuthMode      oauth.AuthMode `envconfig:"APP_SELF_REGISTER_OAUTH_MODE,default=oauth-mtls"`
	OauthTokenPath string         `envconfig:"APP_SELF_REGISTER_OAUTH_TOKEN_PATH,optional"`

	SkipSSLValidation bool `envconfig:"APP_SELF_REGISTER_SKIP_SSL_VALIDATION,default=false"`

	ClientTimeout time.Duration `envconfig:"default=30s"`

	Instances      string                    `envconfig:"APP_SELF_REGISTER_AGGREGATED_XSUAA_INSTANCES"`
	RegionToConfig map[string]InstanceConfig `envconfig:"-"`
}

type InstanceConfig struct {
	ClientID     string
	ClientSecret string
	URL          string
	TokenURL     string
	Cert         string
	Key          string
}

func (i *InstanceConfig) UnmarshalJSON(data []byte) error {
	fmt.Println("TEST")

	instanceBinding := string(data)
	i.ClientID = gjson.Get(instanceBinding, "x509.credentials.clientid").String()
	i.ClientSecret = gjson.Get(instanceBinding, "credentials.clientsecret").String()
	i.URL = gjson.Get(instanceBinding, "x509.credentials.url").String()
	i.TokenURL = gjson.Get(instanceBinding, "x509.credentials.certurl").String()
	i.Cert = gjson.Get(instanceBinding, "x509.credentials.certificate").String()
	i.Key = gjson.Get(instanceBinding, "x509.credentials.key").String()
	return nil
}


type CallerProvider struct {
}

func (c *CallerProvider) GetCaller(config SelfRegConfig, region string) (*securehttp.Caller, error) {
	instanceConfig, exists := config.RegionToConfig[region]
	if !exists {
		return nil, errors.New(fmt.Sprintf("missing configuration for region: %s", region))
	}

	var credentials authpkg.Credentials
	if config.OAuthMode == oauth.Standard {
		credentials = &authpkg.OAuthCredentials{
			ClientID:     instanceConfig.ClientID,
			ClientSecret: instanceConfig.ClientSecret,
			TokenURL:     instanceConfig.URL + config.OauthTokenPath,
		}
	} else if config.OAuthMode == oauth.Mtls {
		mtlsCredentials, err := authpkg.NewOAuthMtlsCredentials(instanceConfig.ClientID, instanceConfig.Cert, instanceConfig.Key, instanceConfig.TokenURL, config.OauthTokenPath)
		if err != nil {
			return nil, errors.Wrap(err, "while creating OAuth Mtls credentials")
		}
		credentials = mtlsCredentials
	} else {
		return nil, errors.New(fmt.Sprintf("unsupported OAuth mode: %s", config.OAuthMode))
	}

	callerConfig := securehttp.CallerConfig{
		Credentials:       credentials,
		ClientTimeout:     config.ClientTimeout,
		SkipSSLValidation: config.SkipSSLValidation,
	}
	caller, err := securehttp.NewCaller(callerConfig)
	if err != nil {
		return nil, err
	}

	return caller, nil
}

type selfRegisterManager struct {
	cfg            SelfRegConfig
	callerProvider ExternalSvcCallerProvider
}

// NewSelfRegisterManager creates a new SelfRegisterManager which is responsible for doing preparation/clean-up during
// self-registration of runtimes configured with values from cfg.
func NewSelfRegisterManager(cfg SelfRegConfig, provider CallerProvider) *selfRegisterManager {
	return &selfRegisterManager{cfg: cfg, callerProvider: &provider}
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
		regionValue, exists := in.Labels["region"]
		if !exists {
			return labels, errors.New("missing region label")
		}

		region := regionValue.(string)

		instanceConfig, exists := s.cfg.RegionToConfig[region]
		if !exists {
			return nil, errors.New(fmt.Sprintf("missing configuration for region: %s", region))
		}

		request, err := s.createSelfRegPrepRequest(id, consumerInfo.ConsumerID, instanceConfig)
		if err != nil {
			return labels, err
		}

		caller, err := s.callerProvider.GetCaller(s.cfg, region)
		if err != nil {
			return nil, err
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

	instanceConfig, exists := s.cfg.RegionToConfig[region]
	if !exists {
		return errors.New(fmt.Sprintf("missing configuration for region: %s", region))
	}

	request, err := s.createSelfRegDelRequest(runtimeID, instanceConfig)
	if err != nil {
		return err
	}

	caller, err := s.callerProvider.GetCaller(s.cfg, region)
	if err != nil {
		return err
	}
	resp, err := caller.Call(request)
	if err != nil {
		return errors.Wrapf(err, "while executing cleanup of self-registered runtime with id %s", runtimeID)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("received unexpected status code %d while cleaning up self-registered runtime with id %s", resp.StatusCode, runtimeID))
	}

	log.C(ctx).Infof("Successfully executed clean-up self-registered runtime with id %s", runtimeID)
	return nil
}

// GetSelfRegDistinguishingLabelKey returns the label key to be used in order to determine whether a runtime
// is being self-registered.
func (s *selfRegisterManager) GetSelfRegDistinguishingLabelKey() string {
	return s.cfg.SelfRegisterDistinguishLabelKey
}

func (s *selfRegisterManager) createSelfRegPrepRequest(runtimeID, tenant string, instanceConfig InstanceConfig) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + runtimeID
	url, err := urlpkg.Parse(instanceConfig.URL)
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

func (s *selfRegisterManager) createSelfRegDelRequest(runtimeID string, instanceConfig InstanceConfig) (*http.Request, error) {
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + runtimeID
	url, err := urlpkg.Parse(instanceConfig.URL)
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
