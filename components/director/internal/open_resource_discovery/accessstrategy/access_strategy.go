package accessstrategy

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
)

// UnsupportedAccessStrategyErr is an error produced when execution of unsupported access strategy takes place.
var UnsupportedAccessStrategyErr = errors.New("unsupported access strategy")

// AccessStrategy is an ORD object
type AccessStrategy struct {
	Type              AccessStrategyType `json:"type"`
	CustomType        AccessStrategyType `json:"customType"`
	CustomDescription string             `json:"customDescription"`
}

// AccessStrategyType represents the possible type of the AccessStrategy
type AccessStrategyType string

// IsSupported checks if the given AccessStrategy is supported by CMP
func (a AccessStrategyType) IsSupported() bool {
	_, ok := supportedAccessStrategies[a]
	return ok
}

func (a AccessStrategyType) Execute(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	if !a.IsSupported() {
		return nil, UnsupportedAccessStrategyErr
	}
	return supportedAccessStrategies[a].Execute(ctx, client, url)
}

const (
	// OpenAccessStrategy is an AccessStrategyType indicating that the ORD document is not secured
	OpenAccessStrategy AccessStrategyType = "open"

	// CMPmTLSAccessStrategy is an AccessStrategyType indicating that the ORD document trusts CMP's client certificate.
	CMPmTLSAccessStrategy AccessStrategyType = "sap:cmp-mtls:v1"

	// CustomAccessStrategy is an AccessStrategyType indicating that not a standard ORD security mechanism is used for the ORD document
	CustomAccessStrategy AccessStrategyType = "custom"
)

var supportedAccessStrategies = map[AccessStrategyType]AccessStrategyExecutor{
	OpenAccessStrategy:    &openAccessStrategyExecutor{},
	CMPmTLSAccessStrategy: newCMPmTLSAccessStrategyExecutor(),
}

// AccessStrategies is a slice of AccessStrategy objects
type AccessStrategies []AccessStrategy

// GetSupported returns the first AccessStrategy in the slice that is supported by CMP
func (as AccessStrategies) GetSupported() (AccessStrategyType, bool) {
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

// AccessStrategyExecutor defines an interface for execution of different access strategies
type AccessStrategyExecutor interface {
	Execute(ctx context.Context, client *http.Client, url string) (*http.Response, error)
}
