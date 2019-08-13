package repo

type NotFound interface {
	IsNotFound() bool
}
type notFoundError struct {
}

func (e *notFoundError) Error() string {
	return "object not found in DB"
}

func (e *notFoundError) IsNotFound() bool {
	return true
}

func IsNotFoundError(err error) bool {
	if _, ok := err.(NotFound); ok {
		return true
	}
	return false
}
