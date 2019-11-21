package persistence

import (
	"github.com/lib/pq"
)

func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	pqerr, ok := err.(*pq.Error)
	if ok {
		if pqerr.Code.Class() == ConstraintViolation {
			return true
		}
	}
	return false
}
