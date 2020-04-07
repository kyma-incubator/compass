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

func (a *AssertionAttributeDeliver) assertionAttributesKeys() map[string]bool {
	attributes := make(map[string]bool, len(a.assertionAttributesTemplate))
	for key, _ := range a.assertionAttributesTemplate {
		attributes[key] = false
	}

	return attributes
}

// GenerateAssertionAttribute remove not existing in template attributes, leaves existing
func (a *AssertionAttributeDeliver) GenerateAssertionAttribute(serviceProvider ServiceProvider) []AssertionAttribute {
	var newAssertionAttributes []AssertionAttribute
	aats := a.assertionAttributesKeys()

	for key, _ := range aats {
		for _, atr := range serviceProvider.AssertionAttributes {
			if atr.AssertionAttribute == key {
				newAssertionAttributes = append(newAssertionAttributes, atr)
				aats[key] = true
			}
		}
	}
	for key, exist := range aats {
		if !exist {
			newAssertionAttributes = append(newAssertionAttributes, a.assertionAttributesTemplate[key])
		}
	}

	return newAssertionAttributes
}
