package authenticator

import (
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
)

type Claims struct {
	Tenant         string                `json:"tenant"`
	ExternalTenant string                `json:"externalTenant"`
	Scopes         string                `json:"scopes"`
	ConsumerID     string                `json:"consumerID"`
	ConsumerType   consumer.ConsumerType `json:"consumerType"`
	Flow           oathkeeper.AuthFlow   `json:"flow"`
	jwt.StandardClaims
}

func (c Claims) Valid() error {
	err := c.StandardClaims.Valid()
	if err != nil {
		return err
	}

	return nil
}
