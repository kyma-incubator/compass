package cert_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubjectExtraction(t *testing.T) {
	for _, testCase := range []struct {
		subject     string
		country     string
		locality    string
		province    string
		org         string
		orgUnit     string
		orgUnits    []string
		uuidOrgUnit string
		commonName  string
	}{
		{
			subject:     "CN=application,OU=OrgUnit,OU=123e4567-e89b-12d3-a456-426614174001,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			country:     "DE",
			locality:    "Waldorf",
			province:    "Waldorf",
			org:         "Org",
			orgUnit:     "OrgUnit",
			orgUnits:    []string{"OrgUnit", "123e4567-e89b-12d3-a456-426614174001"},
			uuidOrgUnit: "123e4567-e89b-12d3-a456-426614174001",
			commonName:  "application",
		},
		{
			subject:     "CN=application,OU=OrgUnit+OU=123e4567-e89b-12d3-a456-426614174001,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			country:     "DE",
			locality:    "Waldorf",
			province:    "Waldorf",
			org:         "Org",
			orgUnit:     "OrgUnit",
			orgUnits:    []string{"OrgUnit", "123e4567-e89b-12d3-a456-426614174001"},
			uuidOrgUnit: "123e4567-e89b-12d3-a456-426614174001",
			commonName:  "application",
		},
		{
			subject:    "CN=,OU=,O=,L=,ST=,C=",
			country:    "",
			locality:   "",
			province:   "",
			org:        "",
			orgUnit:    "",
			orgUnits:   []string{},
			commonName: "",
		},
	} {
		t.Run("should extract subject values", func(t *testing.T) {
			assert.Equal(t, testCase.country, cert.GetCountry(testCase.subject))
			assert.Equal(t, testCase.locality, cert.GetLocality(testCase.subject))
			assert.Equal(t, testCase.province, cert.GetProvince(testCase.subject))
			assert.Equal(t, testCase.org, cert.GetOrganization(testCase.subject))
			assert.Equal(t, testCase.orgUnit, cert.GetOrganizationalUnit(testCase.subject))
			assert.Equal(t, testCase.orgUnits, cert.GetAllOrganizationalUnits(testCase.subject))
			assert.Equal(t, testCase.uuidOrgUnit, cert.GetUUIDOrganizationalUnit(testCase.subject))
			assert.Equal(t, testCase.commonName, cert.GetCommonName(testCase.subject))
		})
	}
}
