package cert_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/stretchr/testify/assert"
)

func TestSubjectExtraction(t *testing.T) {
	for _, testCase := range []struct {
		subject          string
		orgUnitPattern   string
		country          string
		locality         string
		province         string
		org              string
		orgUnit          string
		orgUnits         []string
		uuidOrgUnit      string
		unmatchedOrgUnit string
		commonName       string
	}{
		{
			subject:          "CN=application,OU=OrgUnit1,OU=OrgUnit2,OU=123e4567-e89b-12d3-a456-426614174001,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			orgUnitPattern:   "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|OrgUnit1|OrgUnit2",
			country:          "DE",
			locality:         "Waldorf",
			province:         "Waldorf",
			org:              "Org",
			orgUnit:          "OrgUnit1",
			orgUnits:         []string{"OrgUnit1", "OrgUnit2", "123e4567-e89b-12d3-a456-426614174001"},
			uuidOrgUnit:      "123e4567-e89b-12d3-a456-426614174001",
			unmatchedOrgUnit: "",
			commonName:       "application",
		},
		{
			subject:          "CN=application,OU=OrgUnit1+OU=123e4567-e89b-12d3-a456-426614174001,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			orgUnitPattern:   "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|OrgUnit1",
			country:          "DE",
			locality:         "Waldorf",
			province:         "Waldorf",
			org:              "Org",
			orgUnit:          "OrgUnit1",
			orgUnits:         []string{"OrgUnit1", "123e4567-e89b-12d3-a456-426614174001"},
			uuidOrgUnit:      "123e4567-e89b-12d3-a456-426614174001",
			unmatchedOrgUnit: "",
			commonName:       "application",
		},
		{
			subject:          "CN=application,OU=OrgUnit1,OU=OrgUnit2,OU=UnmatchedOrgUnit,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			orgUnitPattern:   "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|OrgUnit1|OrgUnit2",
			country:          "DE",
			locality:         "Waldorf",
			province:         "Waldorf",
			org:              "Org",
			orgUnit:          "OrgUnit1",
			orgUnits:         []string{"OrgUnit1", "OrgUnit2", "UnmatchedOrgUnit"},
			uuidOrgUnit:      "",
			unmatchedOrgUnit: "UnmatchedOrgUnit",
			commonName:       "application",
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
			assert.Equal(t, testCase.unmatchedOrgUnit, cert.GetUnmatchedOrganizationalUnit(testCase.orgUnitPattern)(testCase.subject))
			assert.Equal(t, testCase.commonName, cert.GetCommonName(testCase.subject))
		})
	}
}
