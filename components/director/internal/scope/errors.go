package scope

import "errors"

var NoScopesInContextError = errors.New("cannot read scopes from context")
var InsufficientScopesError = errors.New("insufficient scopes provided")
var RequiredScopesNotDefinedError = errors.New("required scopes are not defined")
