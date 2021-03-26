/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package osb

import (
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type BrokerOperationType string

const (
	ProvisionOp   BrokerOperationType = "provision_operation"
	BindOp        BrokerOperationType = "bind_operation"
	UnbindOp      BrokerOperationType = "unbind_operation"
	DeprovisionOp BrokerOperationType = "deprovision_operation"
)

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	cause := errors.Cause(err)

	nfe, ok := cause.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}

func IsSucceeded(status *schema.BundleInstanceAuthStatus) bool {
	if status == nil {
		return false
	}
	if status.Condition == schema.BundleInstanceAuthStatusConditionSucceeded {
		return true
	}
	return false
}

func IsFailed(status *schema.BundleInstanceAuthStatus) bool {
	if status == nil {
		return false
	}
	if status.Condition == schema.BundleInstanceAuthStatusConditionFailed {
		return true
	}
	return false
}

func IsInProgress(status *schema.BundleInstanceAuthStatus) bool {
	if status == nil {
		return false
	}
	if status.Condition == schema.BundleInstanceAuthStatusConditionPending {
		return true
	}
	return false
}

func IsUnused(status *schema.BundleInstanceAuthStatus) bool {
	if status == nil {
		return false
	}
	if status.Condition == schema.BundleInstanceAuthStatusConditionUnused {
		return true
	}
	return false
}

type BindingCredentials struct {
	ID          string            `json:"id"`
	Type        AuthType          `json:"credentials_type"`
	TargetURLs  map[string]string `json:"target_urls"`
	AuthDetails AuthDetails       `json:"auth_details"`
}

// AuthType determines the secret structure
type AuthType string

const (
	Undefined   AuthType = ""
	NoAuth      AuthType = "no_auth"
	Oauth       AuthType = "oauth"
	Basic       AuthType = "basic_auth"
	Certificate AuthType = "certificate"
)

type AuthDetails struct {
	RequestParameters *RequestParameters `json:"request_parameters,omitempty"`
	CSRFConfig        *CSRFConfig        `json:"csrf_config,omitempty"`
	Credentials       Auth               `json:"auth,omitempty"`
}

type CSRFConfig struct {
	TokenURL string `json:"token_url"`
}

type Auth interface {
	ToCredentials() *Credentials
}

type NoAuthConfig struct{}

func (oc NoAuthConfig) ToCredentials() *Credentials {
	return nil
}

type OauthConfig struct {
	ClientId          string            `json:"clientId"`
	ClientSecret      string            `json:"clientSecret"`
	TokenURL          string            `json:"tokenUrl"`
	RequestParameters RequestParameters `json:"requestParameters,omitempty"`
}

func (oc OauthConfig) ToCredentials() *Credentials {
	return &Credentials{
		OAuth: &OAuth{
			URL:               oc.TokenURL,
			ClientID:          oc.ClientId,
			ClientSecret:      oc.ClientSecret,
			RequestParameters: &oc.RequestParameters,
		},
	}
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (bc BasicAuthConfig) ToCredentials() *Credentials {
	return &Credentials{
		BasicAuth: &BasicAuth{
			Username: bc.Username,
			Password: bc.Password,
		},
	}
}

type CertificateConfig struct {
	Certificate []byte `json:"certificate"`
	PrivateKey  []byte `json:"privateKey"`
}

func (cc CertificateConfig) ToCredentials() *Credentials {
	return &Credentials{
		CertificateGen: &CertificateGen{
			PrivateKey:  cc.PrivateKey,
			Certificate: cc.Certificate,
		},
	}
}

// Credentials contains OAuth or BasicAuth configuration.
type Credentials struct {
	// OAuth is OAuth configuration.
	OAuth *OAuth
	// BasicAuth is BasicAuth configuration.
	BasicAuth *BasicAuth
	// CertificateGen is CertificateGen configuration.
	CertificateGen *CertificateGen
	// CSRFTokenEndpointURL (optional) to fetch CSRF token
	// Deprecated: This field is only used for old implementation of fetching credentials from Application and Secrets. It is not used by authorization package.
	// It should be removed when it is no longer supported
	CSRFTokenEndpointURL string
}

// BasicAuth contains details of BasicAuth Auth configuration
type BasicAuth struct {
	// Username to use for authentication
	Username string
	// Password to use for authentication
	Password string
}

// OAuth contains details of OAuth configuration
type OAuth struct {
	// URL to OAuth token provider.
	URL string
	// ClientID to use for
	ClientID string
	// ClientSecret to use for
	ClientSecret string
	// RequestParameters will be used with request send by the Application Gateway.
	RequestParameters *RequestParameters
}

// CertificateGen details of CertificateGen configuration
type CertificateGen struct {
	// CommonName of the certificate
	// Deprecated: This field is only used for old implementation of fetching credentials from Application and Secrets
	// It should be removed when it is no longer supported
	CommonName string
	// Certificate generated by Application Registry
	Certificate []byte
	// PrivateKey generated by Application Registry
	PrivateKey []byte
}

// RequestParameters contains Headers and QueryParameters
type RequestParameters struct {
	Headers         *map[string][]string `json:"headers,omitempty"`
	QueryParameters *map[string][]string `json:"query_parameters,omitempty"`
}

func (rp *RequestParameters) unpack() (*map[string][]string, *map[string][]string) {
	if rp == nil {
		return nil, nil
	}
	return rp.Headers, rp.QueryParameters
}

func mapBundleInstanceAuthToModel(instanceAuth schema.BundleInstanceAuth, targets map[string]string) (BindingCredentials, error) {
	var (
		auth = instanceAuth.Auth
		cfg  = AuthDetails{}
	)

	if auth.RequestAuth != nil && auth.RequestAuth.Csrf != nil {
		cfg.CSRFConfig = &CSRFConfig{TokenURL: auth.RequestAuth.Csrf.TokenEndpointURL}
	}

	reqParams, err := resolveAdditionalRequestParameters(auth)
	if err != nil {
		return BindingCredentials{}, errors.Wrapf(err, "while resolving additional request parameters")
	}
	if reqParams.QueryParameters != nil || reqParams.Headers != nil {
		cfg.RequestParameters = reqParams
	}

	var credType AuthType
	switch c := auth.Credential.(type) {
	case nil:
		credType = NoAuth
	case *schema.OAuthCredentialData:
		if c == nil {
			credType = NoAuth
		} else {
			credType = Oauth
			cfg.Credentials = OauthConfig{
				ClientId:     c.ClientID,
				ClientSecret: c.ClientSecret,
				TokenURL:     c.URL,
			}
		}
	case *schema.BasicCredentialData:
		if c == nil {
			credType = NoAuth
		} else {
			credType = Basic
			cfg.Credentials = BasicAuthConfig{
				Username: c.Username,
				Password: c.Password,
			}
		}
	default:
		return BindingCredentials{}, errors.Errorf("got unknown credential type %T", c)
	}

	return BindingCredentials{
		ID:          instanceAuth.ID,
		Type:        credType,
		TargetURLs:  targets,
		AuthDetails: cfg,
	}, nil
}

func resolveAdditionalRequestParameters(auth *schema.Auth) (*RequestParameters, error) {
	additionalRequestParameters := &RequestParameters{}

	if auth.AdditionalHeadersSerialized != nil {
		hed, err := auth.AdditionalHeadersSerialized.Unmarshal()
		if err != nil {
			return nil, err
		}
		additionalRequestParameters.Headers = &hed
	} else {
		if auth.AdditionalHeaders != nil {
			headers := (map[string][]string)(auth.AdditionalHeaders)
			additionalRequestParameters.Headers = &headers
		}
	}

	if auth.AdditionalQueryParamsSerialized != nil {
		params, err := auth.AdditionalQueryParamsSerialized.Unmarshal()
		if err != nil {
			return nil, err
		}
		additionalRequestParameters.QueryParameters = &params
	} else {
		if auth.AdditionalQueryParams != nil {
			qp := (map[string][]string)(auth.AdditionalQueryParams)
			additionalRequestParameters.QueryParameters = &qp
		}
	}

	return additionalRequestParameters, nil
}
