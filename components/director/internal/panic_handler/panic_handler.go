package panic_handler

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

func RecoverFn(ctx context.Context, err interface{}) error {
	errText := fmt.Sprintf("%+v", err)

	return apperrors.NewInternalError(errText)
}
