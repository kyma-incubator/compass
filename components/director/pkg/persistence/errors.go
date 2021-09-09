package persistence

import (
	"github.com/lib/pq"
)

// IsConstraintViolation missing godoc
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code.Class() == ConstraintViolation {
			return true
		}
	}
	return false
}
