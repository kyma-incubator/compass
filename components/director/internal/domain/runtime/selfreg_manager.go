package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/secure_http"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type SelfRegConfig struct {
	SelfRegisterLabelKey            string `envconfig:"default=cmp-xsappname"`
	SelfRegisterLabelValuePrefix    string `envconfig:"default=clone-cmp-"`
	SelfRegisterResponseKey         string `envconfig:"default=appid"`
	SelfRegisterDistinguishLabelKey string `envconfig:"default=xsappname"`
	ClientID                        string `envconfig:"APP_RUNTIME_SVC_CLIENT_ID"`
	ClientSecret                    string `envconfig:"APP_RUNTIME_SVC_CLIENT_SECRET"`
	URL                             string `envconfig:"APP_RUNTIME_SVC_TOKEN_URL"`
	SelfRegisterPath                string
	SelfRegisterNameQueryParam      string
	SelfRegisterTenantQueryParam    string
	SelfRegisterRequestBodyPattern  string
}

type selfRegisterManager struct {
	cfg    SelfRegConfig
	caller *secure_http.Caller
}

func NewSelfRegisterManager(cfg SelfRegConfig) *selfRegisterManager {
	caller, _ := secure_http.NewCaller(&graphql.OAuthCredentialData{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		URL:          cfg.URL + oauthTokenPath,
	})
	return &selfRegisterManager{cfg: cfg, caller: caller}
}

//PrepareRuntimeForSelfRegistration executes the prerequisite calls for self-registration in case the runtime is being self-registered
func (s *selfRegisterManager) PrepareRuntimeForSelfRegistration(ctx context.Context, in *graphql.RuntimeInput) error {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading consumer")
	}
	if distinguishLabel, exists := in.Labels[s.cfg.SelfRegisterDistinguishLabelKey]; exists && consumerInfo.Flow.IsCertFlow() { //this means that the runtime is being self-registered
		distinguishLabelVal, _ := str.Cast(distinguishLabel)

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
			return errors.New(fmt.Sprintf("recieved unexpected status %d while preparing self-registered runtime: %s", response.StatusCode, string(respBytes)))
		}

		selfRegLabelVal := gjson.GetBytes(respBytes, s.cfg.SelfRegisterResponseKey)
		in.Labels[s.cfg.SelfRegisterLabelKey] = selfRegLabelVal.Str

		log.C(ctx).Infof("successfully executed prep for self-registered runtime with label value %s", selfRegLabelVal.Str)
	}
	return nil
}

//CleanupSelfRegisteredRuntime executes cleanup calls for self-registered runtimes
func (s *selfRegisterManager) CleanupSelfRegisteredRuntime(ctx context.Context, selfRegisterDistinguishLabelValue string) error {
	if selfRegisterDistinguishLabelValue != "" {
		request, err := s.createSelfRegDelRequest(selfRegisterDistinguishLabelValue)
		if err != nil {
			return err
		}
		resp, err := s.caller.Call(request)
		if err != nil {
			return errors.Wrapf(err, "while executing cleanup of self-registered runtime with label value %s", selfRegisterDistinguishLabelValue)
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(fmt.Sprintf("recieved unexpected status code %d while cleaning up self-registered runtime with label value %s", resp.StatusCode, selfRegisterDistinguishLabelValue))
		}
		log.C(ctx).Infof("Successfully executed clean-up self-registered runtime with label value %s", selfRegisterDistinguishLabelValue)
	}
	return nil
}

func (s *selfRegisterManager) GetSelfRegDistinguishingLabelKey() string {
	return s.cfg.SelfRegisterDistinguishLabelKey
}

func (s *selfRegisterManager) createSelfRegPrepRequest(distinguishingVal, tenant string) (*http.Request, error) {
	url, err := urlpkg.Parse(s.cfg.URL + s.cfg.SelfRegisterPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating url for preparation of self-registered runtime")
	}
	selfRegLabelVal := s.cfg.SelfRegisterLabelValuePrefix + distinguishingVal
	url.Query().Add(s.cfg.SelfRegisterNameQueryParam, selfRegLabelVal)
	url.Query().Add(s.cfg.SelfRegisterTenantQueryParam, tenant)

	request, err := http.NewRequest(http.MethodPost, url.String(), strings.NewReader(fmt.Sprintf(s.cfg.SelfRegisterRequestBodyPattern, distinguishingVal)))
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
