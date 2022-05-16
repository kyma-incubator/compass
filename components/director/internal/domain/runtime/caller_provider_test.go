package runtime_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/securehttp"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCallerProvider_GetCaller(t *testing.T) {
	var (
		firstRegion        = "eu-1"
		firstClientID      = "client-id"
		firstClientSecret  = "client-secret"
		firstTokenURL      = "token-url"
		secondRegion       = "eu-2"
		secondClientID     = "client-id-2"
		secondClientSecret = "client-secret-2"
		secondTokenURL     = "token-url-2"
		tokenPath          = "/oauth/token"
		timeout            = 15 * time.Second
	)

	const (
		certificate = "-----BEGIN CERTIFICATE-----\nMIIDbjCCAlYCCQDg7pmtw8dIVTANBgkqhkiG9w0BAQsFADB5MQswCQYDVQQGEwJC\nRzENMAsGA1UECAwEVGVzdDENMAsGA1UEBwwEVGVzdDENMAsGA1UECgwEVGVzdDEN\nMAsGA1UECwwEVGVzdDENMAsGA1UEAwwEVGVzdDEfMB0GCSqGSIb3DQEJARYQdGVz\ndEBleGFtcGxlLmNvbTAeFw0yMjAxMjQxMTM4MDFaFw0zMjAxMjIxMTM4MDFaMHkx\nCzAJBgNVBAYTAkJHMQ0wCwYDVQQIDARUZXN0MQ0wCwYDVQQHDARUZXN0MQ0wCwYD\nVQQKDARUZXN0MQ0wCwYDVQQLDARUZXN0MQ0wCwYDVQQDDARUZXN0MR8wHQYJKoZI\nhvcNAQkBFhB0ZXN0QGV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A\nMIIBCgKCAQEAuiFt98GUVTDSCHsOlBcblvUB/02uEmsalsG+DKEufzIVrp4DCxsA\nEsIN85Ywkd1Fsl0vwg9+3ibQlf1XtyXqJ6/jwm2zFdJPM3u2JfGGiiQpscHYp5hS\nlVscBjxZh1CQMKeBXltDsD64EV+XgHGN1aaw9mWKb6iSKsHLhBz594jYMFCnP3wH\nw9/hm6zBAhoF4Xr6UMOp4ZzzY8nzLCGPQuQ9UGp4lyAethrBpsqI6zAxjPKlqhmx\nL3591wkQgTzuL9th54yLEmyEvPTE26ONJBKylH2BqbAFiZPrwet0+PRJSflAfMU8\nYHqqo2AkaY1lmMAZiKDhj1RxMe/jt3HmVQIDAQABMA0GCSqGSIb3DQEBCwUAA4IB\nAQBx8BRhJ59UA3JDL+FHNKwIpxFewxjJwIGWqJTsOh4+rjPK3QeSnF0vt4cnLrCY\n+FLuhhUdFxjeFqJtWN7tHDK3ywSn/yZQTD5Nwcy/F1RmLjl91hjudxO/VewznOlq\nHJlDoM7kW9kOG6xS2HbbSaC1CzU33E90QOwcyCoeVXJ8aMDe6v/kWC65RoI9evg5\n2OxoARA8fpjyUphMTXuVNVI1kd2Uskpo8PePbc1h3OJVzYPIQ4+qMGsu7n3ZdwzI\nqDs2kdBD77k6cBQS+n7g5ETwv5OAgl5q1O17ye/YFNA/T3FhL9to6Nmrkqt7rlnF\nL8uAkeTGuHEATjmosQWUmbYi\n-----END CERTIFICATE-----\n"
		key         = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAuiFt98GUVTDSCHsOlBcblvUB/02uEmsalsG+DKEufzIVrp4D\nCxsAEsIN85Ywkd1Fsl0vwg9+3ibQlf1XtyXqJ6/jwm2zFdJPM3u2JfGGiiQpscHY\np5hSlVscBjxZh1CQMKeBXltDsD64EV+XgHGN1aaw9mWKb6iSKsHLhBz594jYMFCn\nP3wHw9/hm6zBAhoF4Xr6UMOp4ZzzY8nzLCGPQuQ9UGp4lyAethrBpsqI6zAxjPKl\nqhmxL3591wkQgTzuL9th54yLEmyEvPTE26ONJBKylH2BqbAFiZPrwet0+PRJSflA\nfMU8YHqqo2AkaY1lmMAZiKDhj1RxMe/jt3HmVQIDAQABAoIBAH+9xa0N6/FzqhIr\n8ltsaID38cD33QnC++KPYRFl5XViOEM5KrmKdEhragvM/dR92gGJtucmn1lzph/q\nWTLXEJbgPh4ID6pgRf79Xos38bAJFZxrf3e2MKdUei1FaeRWRD9AFqddV100DjvO\nMTnztPX2iujv00zCkl5J1pT7FgrtcYgDPxXQK7dIcHrc9bV9fdTQUnpbVIs/9U7a\n7Qk/eJnEkezbjQCk7+Pgt3ymR29s4vJvyPen3jek0FKhQCxAg6iA5ZOtY+J5AS9e\n3ozZLUEa3b0eOABMw8QnKMtGTmIhLbf9JhISK2Ltsisc/yHHH3KfFE2nayqjvLZf\n5GR62hkCgYEA612EgoRHg4+BSfPfLNG3xsSnM+a98nZOmyxgZ3eNFWpSvi+7MemL\nCJHpwwje412OU1wCc2MtWYvGFY+heL62FxT8+JJLntykZcTQzQoHX3wvaMwopWRi\nJdrv3tEDtSJo9za54kfrNqnVyaxu82r7zgxVbcNiAVR+n7cRXuov288CgYEAynLm\nVI7cIKBOM6U44unkKyIS99Bh57FPjE1QAIsEOiNCWZay4qmzdEboOXjtC95Qyyxn\nTb+MONybwXKkGiLZQZQ2SlgjtEMBDQ+ofk2fK+yHWf4VeLtYWJdBESaAz85xGCCY\nYqlqbFEQd8cl86gTne+emLXp8KrDMuXhbbPvMJsCgYEAgBISAacS9t6GfoQqA0xW\nkNz/EnnTD/UaTst15bci2O1S+tQkK0OmeNJU/eB80AFfabKeTsU/rwMklSTjuz0i\n/ipYgLWyWk47UnknGPsFCgscDQ1SbLTTxz972KWpO83uid6IhT2XGtaNU0D12pRz\nUipZ7fEsCgc9I5FM7XXG9vcCgYBp6xN2ygeBSl2fx6GrlpM5veoOnYeboLjtvsVM\ng28Cu8/K731H+WFaRH7bEtlyjC3ZHrItiznhxgn3e/M/eVwRY2nEG7kSZrv2CWsu\nKY5NfMKT4st5Dwt5zijMwEhEcM3awbL4a4qygPcMs7S3dghNaUCgxQxQTgcyafM3\nYhySYQKBgF7pqQW7ESo1Mp9by+HzJBJsSju5zPBrCZrx8rFAMLCk1uDAIRcUuQtq\n+YwKU8ViemkOHWfN6bePap3/kdVHUxj2xJ6xTAUYHpVOQVMhTw1UmOikiV4FwUo+\nGb5Nk5evWBGhsl2LFqoOqhvFpjftv8+qgRHxmWtj4EoJYWng+hRz\n-----END RSA PRIVATE KEY-----\n"
	)

	firstCallerCreds, err := auth.NewOAuthMtlsCredentials(firstClientID, certificate, key, firstTokenURL, tokenPath)
	require.NoError(t, err)

	firstExpectedCallerCfg := securehttp.CallerConfig{
		Credentials:       firstCallerCreds,
		ClientTimeout:     timeout,
		SkipSSLValidation: false,
	}
	firstExpectedCaller, err := securehttp.NewCaller(firstExpectedCallerCfg)
	require.NoError(t, err)

	cfg := config.SelfRegConfig{
		OAuthMode:         oauth.Mtls,
		OauthTokenPath:    tokenPath,
		SkipSSLValidation: false,
		ClientTimeout:     timeout,
		RegionToInstanceConfig: map[string]config.InstanceConfig{
			firstRegion: {
				ClientID:     firstClientID,
				ClientSecret: firstClientSecret,
				URL:          "url",
				TokenURL:     firstTokenURL,
				Cert:         certificate,
				Key:          key,
			},
			secondRegion: {
				ClientID:     secondClientID,
				ClientSecret: secondClientSecret,
				URL:          "url",
				TokenURL:     secondTokenURL,
				Cert:         certificate,
				Key:          key,
			}},
	}

	testCases := []struct {
		Name                      string
		Config                    config.SelfRegConfig
		Region                    string
		ExpectedExternalSvcCaller runtime.ExternalSvcCaller
		ExpectedErr               error
	}{
		{
			Name:                      "Success",
			Config:                    cfg,
			Region:                    firstRegion,
			ExpectedExternalSvcCaller: firstExpectedCaller,
			ExpectedErr:               nil,
		},
		{
			Name:                      "Returns error when region is missing in the config",
			Config:                    cfg,
			Region:                    "fake-region",
			ExpectedExternalSvcCaller: nil,
			ExpectedErr:               errors.New("missing configuration for region: fake-region"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			c := &runtime.CallerProvider{}
			actualCaller, err := c.GetCaller(testCase.Config, testCase.Region)

			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				ac, ok := actualCaller.(*securehttp.Caller)
				require.True(t, ok)
				ec, ok := testCase.ExpectedExternalSvcCaller.(*securehttp.Caller)
				require.True(t, ok)

				require.Equal(t, ec.Credentials, ac.Credentials)
				require.Equal(t, ec.Provider.Name(), ac.Provider.Name())
			}
		})
	}
}
