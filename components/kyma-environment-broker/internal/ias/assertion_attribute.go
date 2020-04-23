package ias

// AssertionAttributeDeliver ensures required AssertionAttributes
// instead remove all and replace by new one, it will remove only not existing in templates
// and leave existing with probably fresher version of user attributes
type AssertionAttributeDeliver struct {
	assertionAttributesTemplate map[string]AssertionAttribute
}

// NewAssertionAttributeDeliver returns new AssertionAttributeDeliver with default attributes template
func NewAssertionAttributeDeliver() *AssertionAttributeDeliver {
	return &AssertionAttributeDeliver{
		assertionAttributesTemplate: map[string]AssertionAttribute{
			"first_name": {
				AssertionAttribute: "first_name",
				UserAttribute:      "firstName",
			},
			"last_name": {
				AssertionAttribute: "last_name",
				UserAttribute:      "lastName",
			},
			"email": {
				AssertionAttribute: "email",
				UserAttribute:      "mail",
			},
			"groups": {
				AssertionAttribute: "groups",
				UserAttribute:      "companyGroups",
			},
		},
	}
}

// GenerateAssertionAttribute remove not existing in template attributes, leaves existing
func (a *AssertionAttributeDeliver) GenerateAssertionAttribute(serviceProvider ServiceProvider) []AssertionAttribute {
	defaults := a.tmplDeepCopy()
	// overrides defaults with given attr
	for _, overrideAtr := range serviceProvider.AssertionAttributes {
		if _, found := defaults[overrideAtr.AssertionAttribute]; found {
			defaults[overrideAtr.AssertionAttribute] = overrideAtr
		}
	}

	// convert to slice
	var newAssertionAttributes []AssertionAttribute
	for _, atr := range defaults {
		newAssertionAttributes = append(newAssertionAttributes, atr)
	}

	return newAssertionAttributes
}

func (a AssertionAttributeDeliver) tmplDeepCopy() map[string]AssertionAttribute {
	cpy := map[string]AssertionAttribute{}
	for k, v := range a.assertionAttributesTemplate {
		cpy[k] = v
	}
	return cpy
}
