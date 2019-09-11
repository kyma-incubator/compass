package oathkeeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectExtraction(t *testing.T) {

	for _, testCase := range []struct {
		subject    string
		country    string
		locality   string
		province   string
		org        string
		orgUnit    string
		commonName string
	}{
		{
			subject:    "CN=application,OU=OrgUnit,O=Org,L=Waldorf,ST=Waldorf,C=DE",
			country:    "DE",
			locality:   "Waldorf",
			province:   "Waldorf",
			org:        "Org",
			orgUnit:    "OrgUnit",
			commonName: "application",
		},
		{
			subject:    "CN=,OU=,O=,L=,ST=,C=",
			country:    "",
			locality:   "",
			province:   "",
			org:        "",
			orgUnit:    "",
			commonName: "",
		},
	} {
		t.Run("should extract subject values", func(t *testing.T) {
			assert.Equal(t, testCase.country, GetCountry(testCase.subject))
			assert.Equal(t, testCase.locality, GetLocality(testCase.subject))
			assert.Equal(t, testCase.province, GetProvince(testCase.subject))
			assert.Equal(t, testCase.org, GetOrganization(testCase.subject))
			assert.Equal(t, testCase.orgUnit, GetOrganizationalUnit(testCase.subject))
			assert.Equal(t, testCase.commonName, GetCommonName(testCase.subject))
		})
	}

}
