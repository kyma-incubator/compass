package jwtbuilder

import (
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/pkg/errors"
)

type ConsumerType string

const (
	RuntimeConsumer           ConsumerType = "Runtime"
	ApplicationConsumer       ConsumerType = "Application"
	IntegrationSystemConsumer ConsumerType = "Integration System"
	UserConsumer              ConsumerType = "Static User"
)

type jwtTokenClaims struct {
	Scopes       string       `json:"scopes"`
	Tenant       string       `json:"tenant"`
	ConsumerID   string       `json:"consumerID,omitempty"`
	ConsumerType ConsumerType `json:"consumerType,omitempty"`
	jwt.StandardClaims
}

type Consumer struct {
	ID   string
	Type ConsumerType
}

// Build constructs a JWT signed token containing the provided tenant, scopes and consumer information as claims
func Build(tenant string, scopes []string, consumer *Consumer) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwtTokenClaims{
		Tenant:       tenant,
		Scopes:       strings.Join(scopes, " "),
		ConsumerID:   consumer.ID,
		ConsumerType: consumer.Type,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", errors.Wrap(err, "while signing token")
	}

	return signedToken, nil
}
