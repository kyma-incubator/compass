package oathkeeper_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, testCase.country, oathkeeper.GetCountry(testCase.subject))
			assert.Equal(t, testCase.locality, oathkeeper.GetLocality(testCase.subject))
			assert.Equal(t, testCase.province, oathkeeper.GetProvince(testCase.subject))
			assert.Equal(t, testCase.org, oathkeeper.GetOrganization(testCase.subject))
			assert.Equal(t, testCase.orgUnit, oathkeeper.GetOrganizationalUnit(testCase.subject))
			assert.Equal(t, testCase.orgUnits, oathkeeper.GetAllOrganizationalUnits(testCase.subject))
			assert.Equal(t, testCase.uuidOrgUnit, oathkeeper.GetUUIDOrganizationalUnit(testCase.subject))
			assert.Equal(t, testCase.commonName, oathkeeper.GetCommonName(testCase.subject))
		})
	}

}
