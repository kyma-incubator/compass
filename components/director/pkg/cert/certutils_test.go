package cert_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
)

func TestSubjectExtraction(t *testing.T) {
	for _, testCase := range []struct {
		subject                string
		orgUnitPattern         string
		orgRegionPattern       string
		country                string
		locality               string
		province               string
		org                    string
		orgUnit                string
		orgUnits               []string
		uuidOrgUnit            string
		remainingOrgUnit       string
		commonName             string
		possibleOrgUnitMatches int
	}{
		{
			subject:                "CN=application,OU=OrgUnit1,OU=OrgUnit2,OU=123e4567-e89b-12d3-a456-426614174001,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			orgUnitPattern:         "OrgUnit1",
			orgRegionPattern:       "OrgUnit2",
			country:                "DE",
			locality:               "Waldorf",
			province:               "Waldorf",
			org:                    "Org",
			orgUnit:                "OrgUnit1",
			orgUnits:               []string{"OrgUnit1", "OrgUnit2", "123e4567-e89b-12d3-a456-426614174001"},
			uuidOrgUnit:            "123e4567-e89b-12d3-a456-426614174001",
			remainingOrgUnit:       "123e4567-e89b-12d3-a456-426614174001",
			commonName:             "application",
			possibleOrgUnitMatches: 2,
		},
		{
			subject:                "CN=application,OU=OrgUnit1,OU=123e4567-e89b-12d3-a456-426614174001,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			orgUnitPattern:         "OrgUnit1",
			country:                "DE",
			locality:               "Waldorf",
			province:               "Waldorf",
			org:                    "Org",
			orgUnit:                "OrgUnit1",
			orgUnits:               []string{"OrgUnit1", "123e4567-e89b-12d3-a456-426614174001"},
			uuidOrgUnit:            "123e4567-e89b-12d3-a456-426614174001",
			remainingOrgUnit:       "123e4567-e89b-12d3-a456-426614174001",
			commonName:             "application",
			possibleOrgUnitMatches: 1,
		},
		{
			subject:                "CN=application,OU=OrgUnit1,OU=OrgUnit2,OU=RemainingOrgUnit,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			orgUnitPattern:         "OrgUnit1",
			orgRegionPattern:       "OrgUnit2",
			country:                "DE",
			locality:               "Waldorf",
			province:               "Waldorf",
			org:                    "Org",
			orgUnit:                "OrgUnit1",
			orgUnits:               []string{"OrgUnit1", "OrgUnit2", "RemainingOrgUnit"},
			uuidOrgUnit:            "",
			remainingOrgUnit:       "RemainingOrgUnit",
			commonName:             "application",
			possibleOrgUnitMatches: 2,
		},
		{
			subject:                "CN=,OU=,O=,L=,ST=,C=",
			orgUnitPattern:         "(OrgUnit1|OrgUnit2|OrgUnit3)|OrgUnit4",
			country:                "",
			locality:               "",
			province:               "",
			org:                    "",
			orgUnit:                "",
			orgUnits:               []string{},
			commonName:             "",
			possibleOrgUnitMatches: 2,
		},
		{
			subject:                "CN=,OU=,O=,L=,ST=,C=",
			country:                "",
			locality:               "",
			province:               "",
			org:                    "",
			orgUnit:                "",
			orgUnits:               []string{},
			commonName:             "",
			possibleOrgUnitMatches: 0,
		},
	} {
		t.Run("should extract subject values", func(t *testing.T) {
			require.Equal(t, testCase.locality, cert.GetLocality(testCase.subject))
			require.Equal(t, testCase.country, cert.GetCountry(testCase.subject))
			require.Equal(t, testCase.province, cert.GetProvince(testCase.subject))
			require.Equal(t, testCase.org, cert.GetOrganization(testCase.subject))
			require.Equal(t, testCase.orgUnit, cert.GetOrganizationalUnit(testCase.subject))
			require.Equal(t, testCase.orgUnits, cert.GetAllOrganizationalUnits(testCase.subject))
			require.Equal(t, testCase.uuidOrgUnit, cert.GetUUIDOrganizationalUnit(testCase.subject))
			require.Equal(t, testCase.remainingOrgUnit, cert.GetRemainingOrganizationalUnit(testCase.orgUnitPattern, testCase.orgRegionPattern)(testCase.subject))
			require.Equal(t, testCase.commonName, cert.GetCommonName(testCase.subject))
			require.Equal(t, testCase.possibleOrgUnitMatches, cert.GetPossibleRegexTopLevelMatches(constructRegex(testCase.orgUnitPattern, testCase.orgRegionPattern)))
		})
	}
}

func constructRegex(patterns ...string) string {
	nonEmptyStr := make([]string, 0)
	for _, pattern := range patterns {
		if len(pattern) > 0 {
			nonEmptyStr = append(nonEmptyStr, pattern)
		}
	}
	return strings.Join(nonEmptyStr, "|")
}

func TestParseCertificate(t *testing.T) {
	const certificate = "-----BEGIN CERTIFICATE-----\nMIIDbjCCAlYCCQDg7pmtw8dIVTANBgkqhkiG9w0BAQsFADB5MQswCQYDVQQGEwJC\nRzENMAsGA1UECAwEVGVzdDENMAsGA1UEBwwEVGVzdDENMAsGA1UECgwEVGVzdDEN\nMAsGA1UECwwEVGVzdDENMAsGA1UEAwwEVGVzdDEfMB0GCSqGSIb3DQEJARYQdGVz\ndEBleGFtcGxlLmNvbTAeFw0yMjAxMjQxMTM4MDFaFw0zMjAxMjIxMTM4MDFaMHkx\nCzAJBgNVBAYTAkJHMQ0wCwYDVQQIDARUZXN0MQ0wCwYDVQQHDARUZXN0MQ0wCwYD\nVQQKDARUZXN0MQ0wCwYDVQQLDARUZXN0MQ0wCwYDVQQDDARUZXN0MR8wHQYJKoZI\nhvcNAQkBFhB0ZXN0QGV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A\nMIIBCgKCAQEAuiFt98GUVTDSCHsOlBcblvUB/02uEmsalsG+DKEufzIVrp4DCxsA\nEsIN85Ywkd1Fsl0vwg9+3ibQlf1XtyXqJ6/jwm2zFdJPM3u2JfGGiiQpscHYp5hS\nlVscBjxZh1CQMKeBXltDsD64EV+XgHGN1aaw9mWKb6iSKsHLhBz594jYMFCnP3wH\nw9/hm6zBAhoF4Xr6UMOp4ZzzY8nzLCGPQuQ9UGp4lyAethrBpsqI6zAxjPKlqhmx\nL3591wkQgTzuL9th54yLEmyEvPTE26ONJBKylH2BqbAFiZPrwet0+PRJSflAfMU8\nYHqqo2AkaY1lmMAZiKDhj1RxMe/jt3HmVQIDAQABMA0GCSqGSIb3DQEBCwUAA4IB\nAQBx8BRhJ59UA3JDL+FHNKwIpxFewxjJwIGWqJTsOh4+rjPK3QeSnF0vt4cnLrCY\n+FLuhhUdFxjeFqJtWN7tHDK3ywSn/yZQTD5Nwcy/F1RmLjl91hjudxO/VewznOlq\nHJlDoM7kW9kOG6xS2HbbSaC1CzU33E90QOwcyCoeVXJ8aMDe6v/kWC65RoI9evg5\n2OxoARA8fpjyUphMTXuVNVI1kd2Uskpo8PePbc1h3OJVzYPIQ4+qMGsu7n3ZdwzI\nqDs2kdBD77k6cBQS+n7g5ETwv5OAgl5q1O17ye/YFNA/T3FhL9to6Nmrkqt7rlnF\nL8uAkeTGuHEATjmosQWUmbYi\n-----END CERTIFICATE-----\n"
	const key = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAuiFt98GUVTDSCHsOlBcblvUB/02uEmsalsG+DKEufzIVrp4D\nCxsAEsIN85Ywkd1Fsl0vwg9+3ibQlf1XtyXqJ6/jwm2zFdJPM3u2JfGGiiQpscHY\np5hSlVscBjxZh1CQMKeBXltDsD64EV+XgHGN1aaw9mWKb6iSKsHLhBz594jYMFCn\nP3wHw9/hm6zBAhoF4Xr6UMOp4ZzzY8nzLCGPQuQ9UGp4lyAethrBpsqI6zAxjPKl\nqhmxL3591wkQgTzuL9th54yLEmyEvPTE26ONJBKylH2BqbAFiZPrwet0+PRJSflA\nfMU8YHqqo2AkaY1lmMAZiKDhj1RxMe/jt3HmVQIDAQABAoIBAH+9xa0N6/FzqhIr\n8ltsaID38cD33QnC++KPYRFl5XViOEM5KrmKdEhragvM/dR92gGJtucmn1lzph/q\nWTLXEJbgPh4ID6pgRf79Xos38bAJFZxrf3e2MKdUei1FaeRWRD9AFqddV100DjvO\nMTnztPX2iujv00zCkl5J1pT7FgrtcYgDPxXQK7dIcHrc9bV9fdTQUnpbVIs/9U7a\n7Qk/eJnEkezbjQCk7+Pgt3ymR29s4vJvyPen3jek0FKhQCxAg6iA5ZOtY+J5AS9e\n3ozZLUEa3b0eOABMw8QnKMtGTmIhLbf9JhISK2Ltsisc/yHHH3KfFE2nayqjvLZf\n5GR62hkCgYEA612EgoRHg4+BSfPfLNG3xsSnM+a98nZOmyxgZ3eNFWpSvi+7MemL\nCJHpwwje412OU1wCc2MtWYvGFY+heL62FxT8+JJLntykZcTQzQoHX3wvaMwopWRi\nJdrv3tEDtSJo9za54kfrNqnVyaxu82r7zgxVbcNiAVR+n7cRXuov288CgYEAynLm\nVI7cIKBOM6U44unkKyIS99Bh57FPjE1QAIsEOiNCWZay4qmzdEboOXjtC95Qyyxn\nTb+MONybwXKkGiLZQZQ2SlgjtEMBDQ+ofk2fK+yHWf4VeLtYWJdBESaAz85xGCCY\nYqlqbFEQd8cl86gTne+emLXp8KrDMuXhbbPvMJsCgYEAgBISAacS9t6GfoQqA0xW\nkNz/EnnTD/UaTst15bci2O1S+tQkK0OmeNJU/eB80AFfabKeTsU/rwMklSTjuz0i\n/ipYgLWyWk47UnknGPsFCgscDQ1SbLTTxz972KWpO83uid6IhT2XGtaNU0D12pRz\nUipZ7fEsCgc9I5FM7XXG9vcCgYBp6xN2ygeBSl2fx6GrlpM5veoOnYeboLjtvsVM\ng28Cu8/K731H+WFaRH7bEtlyjC3ZHrItiznhxgn3e/M/eVwRY2nEG7kSZrv2CWsu\nKY5NfMKT4st5Dwt5zijMwEhEcM3awbL4a4qygPcMs7S3dghNaUCgxQxQTgcyafM3\nYhySYQKBgF7pqQW7ESo1Mp9by+HzJBJsSju5zPBrCZrx8rFAMLCk1uDAIRcUuQtq\n+YwKU8ViemkOHWfN6bePap3/kdVHUxj2xJ6xTAUYHpVOQVMhTw1UmOikiV4FwUo+\nGb5Nk5evWBGhsl2LFqoOqhvFpjftv8+qgRHxmWtj4EoJYWng+hRz\n-----END RSA PRIVATE KEY-----\n"

	t.Run("non-base64-encoded-cert", func(t *testing.T) {
		_, err := cert.ParseCertificate(certificate, key)
		require.NoError(t, err, "failed to parse certificate")
	})

	t.Run("base64-encoded-cert", func(t *testing.T) {
		base64EncodedCert := base64.StdEncoding.EncodeToString([]byte(certificate))
		base64EncodedKey := base64.StdEncoding.EncodeToString([]byte(key))

		_, err := cert.ParseCertificate(base64EncodedCert, base64EncodedKey)
		require.NoError(t, err, "failed to parse certificate")
	})
}
