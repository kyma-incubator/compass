package httputils

import (
	"context"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

func Close(ctx context.Context, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.C(ctx).WithError(err).Warnf("Warning: failed to close")
	}
}
