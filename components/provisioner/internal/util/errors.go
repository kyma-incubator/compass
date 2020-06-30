package util

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/apperrors"
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
