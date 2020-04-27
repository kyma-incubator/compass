package ias

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertionAttributeDeliver_GenerateAssertionAttribute(t *testing.T) {
	// given
	attr := NewAssertionAttributeDeliver()

	// when
	attributes := attr.GenerateAssertionAttribute(ServiceProvider{
		AssertionAttributes: []AssertionAttribute{
			{
				AssertionAttribute: "last_name",
				UserAttribute:      "lastName",
			},
			{
				AssertionAttribute: "should_be_removed",
				UserAttribute:      "shouldBeRemoved",
			},
		},
	})

	// then
	assert.ElementsMatch(t, []AssertionAttribute{
		{
			AssertionAttribute: "first_name",
			UserAttribute:      "firstName",
		},
		{
			AssertionAttribute: "last_name",
			UserAttribute:      "lastName",
		},
		{
			AssertionAttribute: "email",
			UserAttribute:      "mail",
		},
		{
			AssertionAttribute: "groups",
			UserAttribute:      "companyGroups",
		},
	}, attributes)
}
