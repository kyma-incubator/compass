package apperrors

import "strings"

const (
	CodeInternal = iota + 1
	CodeNotFound
	CodeAlreadyExists
	CodeWrongInput
	CodeUpstreamServerCallFailed
	CodeForbidden
)

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Object was not found")
}
