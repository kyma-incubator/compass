package httputils

import (
	"context"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

func Close(ctx context.Context, closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.C(ctx).WithError(err).Warnf("Warning: failed to close")
	}
}
