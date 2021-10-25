package accessstrategy

import (
	"context"
	"net/http"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/pkg/errors"
)

var supportedAccessStrategies = map[Type]Executor{
	OpenAccessStrategy:    &openAccessStrategyExecutor{},
	CMPmTLSAccessStrategy: NewCMPmTLSAccessStrategyExecutor(),
}

// UnsupportedErr is an error produced when execution of unsupported access strategy takes place.
var UnsupportedErr = errors.New("unsupported access strategy")

// AccessStrategy is an ORD object
type AccessStrategy struct {
	Type              Type   `json:"type"`
	CustomType        Type   `json:"customType"`
	CustomDescription string `json:"customDescription"`
}

// Validate validates a given access strategy
func (as AccessStrategy) Validate() error {
	const CustomTypeRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	return validation.ValidateStruct(&as,
		validation.Field(&as.Type, validation.Required, validation.In(OpenAccessStrategy, CMPmTLSAccessStrategy, CustomAccessStrategy), validation.When(as.CustomType != "", validation.In("custom"))),
		validation.Field(&as.CustomType, validation.When(as.CustomType != "", validation.Match(regexp.MustCompile(CustomTypeRegex)))),
		validation.Field(&as.CustomDescription, validation.When(as.Type != "custom", validation.Empty)),
	)
}

// Type represents the possible type of the AccessStrategy
type Type string

// IsSupported checks if the given AccessStrategy is supported by CMP
func (a Type) IsSupported() bool {
	_, ok := supportedAccessStrategies[a]
	return ok
}

const (
	// OpenAccessStrategy is an AccessStrategyType indicating that the ORD document is not secured
	OpenAccessStrategy Type = "open"

	// CMPmTLSAccessStrategy is an AccessStrategyType indicating that the ORD document trusts CMP's client certificate.
	CMPmTLSAccessStrategy Type = "sap:cmp-mtls:v1"

	// CustomAccessStrategy is an AccessStrategyType indicating that not a standard ORD security mechanism is used for the ORD document
	CustomAccessStrategy Type = "custom"
)

// AccessStrategies is a slice of AccessStrategy objects
type AccessStrategies []AccessStrategy

// GetSupported returns the first AccessStrategy in the slice that is supported by CMP
func (as AccessStrategies) GetSupported() (Type, bool) {
	for _, v := range as {
		if v.Type.IsSupported() {
			return v.Type, true
		}
		if v.Type == CustomAccessStrategy && v.CustomType.IsSupported() {
			return v.CustomType, true
		}
	}
	return "", false
}

// Executor defines an interface for execution of different access strategies
//go:generate mockery --name=Executor --output=automock --outpkg=automock --case=underscore
type Executor interface {
	Execute(ctx context.Context, client *http.Client, url string) (*http.Response, error)
}

// ExecutorProvider defines an interface for access strategy executor provider
//go:generate mockery --name=ExecutorProvider --output=automock --outpkg=automock --case=underscore
type ExecutorProvider interface {
	Provide(accessStrategyType Type) (Executor, error)
}
