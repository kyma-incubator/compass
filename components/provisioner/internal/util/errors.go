package util

import (
	"errors"
	"testing"

	"github.com/kyma-project/control-plane/components/provisioner/internal/apperrors"
	"gotest.tools/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func K8SErrorToAppError(err error) apperrors.AppError {
	if k8serrors.IsBadRequest(err) {
		return apperrors.BadRequest(err.Error())
	}
	if k8serrors.IsForbidden(err) {
		return apperrors.Forbidden(err.Error())
	}
	return apperrors.Internal(err.Error())
}

func CheckErrorType(t *testing.T, err error, errCode apperrors.ErrCode) {
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fail()
	}
	assert.Equal(t, appErr.Code(), errCode)
}
