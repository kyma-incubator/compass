package repo

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/lib/pq"
)

func specificPqError(err *pq.Error) error {
	if err.Code.Class() == "23" {
		return apperrors.NewConstraintViolationError(err.Table)
	}
	return nil
}
