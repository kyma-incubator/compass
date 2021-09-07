package panichandler

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

// RecoverFn missing godoc
func RecoverFn(ctx context.Context, err interface{}) error {
	errText := fmt.Sprintf("%+v", err)

	return apperrors.NewInternalError(errText)
}
