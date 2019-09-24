package scope

import (
	"errors"
	"fmt"
)

var NoScopesInContextError = errors.New("cannot read scopes from context")
var RequiredScopesNotDefinedError = errors.New("required scopes are not defined")

func InsufficientScopesError(required, actual []string) error {
	return fmt.Errorf("insufficient scopes provided, required: %v, actual: %v", required, actual)
}
